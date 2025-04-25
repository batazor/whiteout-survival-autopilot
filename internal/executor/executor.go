package executor

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/metrics"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
	"github.com/batazor/whiteout-survival-autopilot/internal/utils"
)

// UseCaseExecutor описывает интерфейс для выполнения UseCase
type UseCaseExecutor interface {
	ExecuteUseCase(ctx context.Context, uc *domain.UseCase, state *domain.Gamer, queue *redis_queue.Queue)
	Analyzer() Analyzer
}

// Analyzer описывает интерфейс для анализа скриншота и обновления состояния игрока
type Analyzer interface {
	AnalyzeAndUpdateState(imagePath string, state *domain.Gamer, rules []domain.AnalyzeRule, queue *redis_queue.Queue) (*domain.Gamer, error)
}

// NewUseCaseExecutor возвращает реализацию UseCaseExecutor
func NewUseCaseExecutor(
	logger *slog.Logger,
	triggerEvaluator config.TriggerEvaluator,
	analyzer Analyzer,
	adb adb.DeviceController,
	area *config.AreaLookup,
	botName string,
	queue *redis_queue.Queue,
) UseCaseExecutor {
	return &executorImpl{
		logger:           logger,
		triggerEvaluator: triggerEvaluator,
		analyzer:         analyzer,
		adb:              adb,
		area:             area,
		botName:          botName,
		queue:            queue,
	}
}

type executorImpl struct {
	logger           *slog.Logger
	triggerEvaluator config.TriggerEvaluator
	analyzer         Analyzer
	adb              adb.DeviceController
	area             *config.AreaLookup
	botName          string
	queue            *redis_queue.Queue
}

func (e *executorImpl) Analyzer() Analyzer {
	return e.analyzer
}

// ExecuteUseCase выполняет сам UseCase целиком
func (e *executorImpl) ExecuteUseCase(ctx context.Context, uc *domain.UseCase, gamer *domain.Gamer, queue *redis_queue.Queue) {
	// Создаём span для всего UseCase
	start := time.Now()
	tracer := otel.Tracer("bot")
	ctx, span := tracer.Start(ctx, uc.Name)
	defer span.End()

	// Извлекаем TraceID для логов
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	// Проверяем триггер UseCase
	if uc.Trigger != "" {
		ok, err := e.triggerEvaluator.EvaluateTrigger(uc.Trigger, gamer)
		if err != nil {
			e.logger.Error("Trigger evaluation failed",
				slog.String("usecase", uc.Name),
				slog.String("trigger", uc.Trigger),
				slog.Any("error", err),
			)
			return
		}

		if !ok {
			e.logger.Warn("Trigger not met, skipping usecase",
				slog.String("usecase", uc.Name),
				slog.String("trigger", uc.Trigger),
			)
			return
		}
	}

	// Логируем старт UseCase с TraceID
	e.logger.Info("=== Start usecase ===",
		slog.String("name", uc.Name),
		slog.String("trace_id", traceID),
	)

	for _, step := range uc.Steps {
		// Вызываем вложенные шаги
		e.runStep(ctx, step, 0, gamer)
	}

	// Логируем окончание UseCase с TraceID
	e.logger.Info("=== End usecase ===",
		slog.String("name", uc.Name),
		slog.String("trace_id", traceID),
	)

	// Если UseCase успешно выполнен — ставим TTL (если есть)
	if uc.TTL > 0 && queue != nil {
		if err := queue.SetLastExecuted(ctx, gamer.ID, uc.Name, uc.TTL); err != nil {
			e.logger.Error("Failed to set last executed TTL", slog.Any("error", err))
		}
	}

	// Счётчики и метрики
	metrics.UsecaseTotal.WithLabelValues(uc.Name).Inc()
	metrics.UsecaseDuration.WithLabelValues(uc.Name).Observe(time.Since(start).Seconds())

	// Пример записи метрик состояния игрока
	if gamer != nil {
		// Сила игрока
		metrics.GamerPowerGauge.WithLabelValues(gamer.Nickname).Set(float64(gamer.Power))

		// Уровень печки (если доступен)
		if gamer.Buildings.Furnace.Level > 0 {
			metrics.GamerFurnaceLevel.WithLabelValues(gamer.Nickname).Set(float64(gamer.Buildings.Furnace.Level))
		}
	}
}

// runStep выполняет один шаг UseCase (возможно рекурсивно вызывает сам себя для вложенных шагов)
func (e *executorImpl) runStep(ctx context.Context, step domain.Step, indent int, gamer *domain.Gamer) bool {
	// Начинаем трейс на каждый шаг
	ctx, stepSpan := otel.Tracer("bot").Start(ctx, "runStep: "+step.Action)
	defer stepSpan.End()

	select {
	case <-ctx.Done():
		e.logger.Warn("Step cancelled by context")
		return true
	default:
	}

	prefix := strings.Repeat("  ", indent)

	// Если есть step.Click — кликаем
	if step.Click != "" {
		e.logger.Info(prefix+"Click", slog.String("target", step.Click))

		err := e.adb.ClickRegion(step.Click, e.area)
		if err != nil {
			e.logger.Error(prefix+"Failed to click region",
				slog.String("target", step.Click),
				slog.Any("error", err),
			)
			return true
		}
	}

	// Если есть step.Action — выполняем её
	if step.Action != "" {
		e.logger.Info(prefix+"Action", slog.String("action", step.Action))

		switch step.Action {

		// Сброс state-поля: "reset"
		case "reset":
			if step.Set == "" {
				e.logger.Warn(prefix + "Reset skipped: missing 'set' field")
				return false
			}

			// Получим текущее значение для логирования
			prevVal, _ := utils.GetStateFieldByPath(gamer, step.Set)

			if err := utils.SetStateFieldByPath(gamer, step.Set, step.To); err != nil {
				e.logger.Error(prefix+"Failed to reset state field",
					slog.String("path", step.Set),
					slog.Any("from", prevVal),
					slog.Any("to", step.To),
					slog.Any("error", err),
				)
			} else {
				e.logger.Info(prefix+"State field reset",
					slog.String("path", step.Set),
					slog.Any("from", prevVal),
					slog.Any("to", step.To),
				)
			}

		// Организация цикла: "loop"
		case "loop":
			if step.Trigger == "" {
				e.logger.Warn(prefix + "Loop trigger is missing, skipping loop")
				return false
			}

			// Создаём отдельный спан на весь цикл
			loopCtx, loopSpan := otel.Tracer("bot").Start(ctx, prefix+"loop: "+step.Trigger)
			defer loopSpan.End()

			e.logger.Info(prefix+"Entering loop", slog.String("trigger", step.Trigger))

			for {
				select {
				case <-loopCtx.Done():
					e.logger.Warn(prefix + "Loop interrupted by context")
					return true
				default:
				}

				shouldContinue, err := e.triggerEvaluator.EvaluateTrigger(step.Trigger, gamer)
				if err != nil {
					e.logger.Error(prefix+"Trigger evaluation failed", slog.Any("error", err))
					break
				}
				if !shouldContinue {
					e.logger.Info(prefix + "Loop trigger returned false, exiting loop")
					break
				}

				for _, s := range step.Steps {
					if stopped := e.runStep(loopCtx, s, indent+1, gamer); stopped {
						e.logger.Info(prefix + "Loop stopped manually (loop_stop)")
						return false
					}
				}
			}

		// Принудительный выход из цикла
		case "loop_stop":
			e.logger.Info(prefix + "Received loop_stop")
			return true

		// Скриншот с последующим анализом
		case "screenshot":
			imagePath := filepath.Join("out", fmt.Sprintf("step_%d.png", indent))
			e.logger.Info(prefix+"Taking screenshot", slog.String("path", imagePath))

			if _, err := e.adb.Screenshot(imagePath); err != nil {
				e.logger.Error(prefix+"Failed to capture screenshot", slog.Any("error", err))
				return false
			}

			// Если есть правила анализа
			if len(step.Analyze) > 0 {
				_, analyzeSpan := otel.Tracer("bot").Start(ctx, prefix+"AnalyzeAndUpdateState")
				defer analyzeSpan.End()

				newState, err := e.analyzer.AnalyzeAndUpdateState(imagePath, gamer, step.Analyze, e.queue)
				if err != nil {
					e.logger.Error(prefix+"Analyze failed", slog.Any("error", err))
				} else {
					*gamer = *newState
					e.logger.Info(prefix + "Analyze completed and state updated")
				}
			}
		}
	}

	// Если есть step.Wait — ждём
	if step.Wait > 0 {
		e.logger.Info(prefix+"Wait", slog.Duration("duration", step.Wait))
		select {
		case <-time.After(step.Wait):
		case <-ctx.Done():
			e.logger.Warn(prefix+"Wait interrupted by context cancel", slog.Duration("wait", step.Wait))
			return true
		}
	}

	// Если есть условие if/then/else
	if step.If != nil {
		// Заводим отдельный спан для всего `if`
		ifCtx, ifSpan := otel.Tracer("bot").Start(ctx, prefix+"if: "+step.If.Trigger)
		defer ifSpan.End()

		e.logger.Info(prefix+"If Trigger", slog.String("expr", step.If.Trigger))

		result, err := e.triggerEvaluator.EvaluateTrigger(step.If.Trigger, gamer)
		if err != nil {
			e.logger.Error(prefix+"Trigger evaluation failed",
				slog.String("expr", step.If.Trigger),
				slog.Any("error", err),
			)
			return false
		}

		if result {
			// then
			thenCtx, thenSpan := otel.Tracer("bot").Start(ifCtx, prefix+"then")
			defer thenSpan.End()

			e.logger.Info(prefix + "Condition met: executing THEN")
			for _, s := range step.If.Then {
				stopped := e.runStep(thenCtx, s, indent+1, gamer)
				if stopped {
					return true
				}
			}
		} else if len(step.If.Else) > 0 {
			// else
			elseCtx, elseSpan := otel.Tracer("bot").Start(ifCtx, prefix+"else")
			defer elseSpan.End()

			e.logger.Info(prefix + "Condition NOT met: executing ELSE")
			for _, s := range step.If.Else {
				stopped := e.runStep(elseCtx, s, indent+1, gamer)
				if stopped {
					return true
				}
			}
		}
	}

	// Длинный тап (longtap)
	if step.Longtap != "" {
		e.logger.Info(prefix+"Longtap", slog.String("target", step.Longtap), slog.Duration("hold", step.Wait))

		bbox, err := e.area.GetRegionByName(step.Longtap)
		if err != nil {
			e.logger.Error(prefix+"Failed to find region for longtap",
				slog.String("target", step.Longtap),
				slog.Any("error", err),
			)
			return true
		}

		x, y, _, _ := bbox.ToPixels()
		err = e.adb.Swipe(x, y, x, y, step.Wait) // свайп на то же место с заданным временем
		if err != nil {
			e.logger.Error(prefix+"Failed to perform longtap",
				slog.String("target", step.Longtap),
				slog.Any("error", err),
			)
			return true
		}
	}

	return false
}
