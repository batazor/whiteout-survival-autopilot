package executor

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/utils"
)

type UseCaseExecutor interface {
	ExecuteUseCase(ctx context.Context, uc *domain.UseCase, state *domain.Gamer)
}

type Analyzer interface {
	AnalyzeAndUpdateState(imagePath string, state *domain.Gamer, rules []domain.AnalyzeRule) (*domain.Gamer, error)
}

func NewUseCaseExecutor(
	logger *slog.Logger,
	triggerEvaluator config.TriggerEvaluator,
	analyzer Analyzer,
	adb adb.DeviceController,
	area *config.AreaLookup,
) UseCaseExecutor {
	return &executorImpl{
		logger:           logger,
		triggerEvaluator: triggerEvaluator,
		analyzer:         analyzer,
		adb:              adb,
		area:             area,
	}
}

type executorImpl struct {
	logger           *slog.Logger
	triggerEvaluator config.TriggerEvaluator
	analyzer         Analyzer
	adb              adb.DeviceController
	area             *config.AreaLookup
}

func (e *executorImpl) ExecuteUseCase(ctx context.Context, uc *domain.UseCase, gamer *domain.Gamer) {
	select {
	case <-ctx.Done():
		e.logger.Warn("Usecase cancelled before execution started",
			slog.String("usecase", uc.Name))
		return
	default:
		// Continue if not cancelled
	}

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

	e.logger.Info("=== Start usecase ===", slog.String("name", uc.Name))
	for _, step := range uc.Steps {
		e.runStep(ctx, step, 0, gamer)
	}
	e.logger.Info("=== End usecase ===", slog.String("name", uc.Name))
}

func (e *executorImpl) runStep(ctx context.Context, step domain.Step, indent int, gamer *domain.Gamer) bool {
	select {
	case <-ctx.Done():
		e.logger.Warn("Step cancelled by context")
		return true
	default:
	}

	prefix := strings.Repeat("  ", indent)

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

	if step.Action != "" {
		e.logger.Info(prefix+"Action", slog.String("action", step.Action))

		switch step.Action {
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

		case "loop":
			if step.Trigger == "" {
				e.logger.Warn(prefix + "Loop trigger is missing, skipping loop")
				return false
			}
			e.logger.Info(prefix+"Entering loop", slog.String("trigger", step.Trigger))

			for {
				select {
				case <-ctx.Done():
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
					if stopped := e.runStep(ctx, s, indent+1, gamer); stopped {
						e.logger.Info(prefix + "Loop stopped manually (loop_stop)")
						return false
					}
				}
			}

		case "loop_stop":
			e.logger.Info(prefix + "Received loop_stop")
			return true

		case "screenshot":
			imagePath := filepath.Join("out", fmt.Sprintf("%s_step_%d.png", strings.ReplaceAll(step.UsecaseName, " ", "_"), indent))
			e.logger.Info(prefix+"Taking screenshot", slog.String("path", imagePath))

			if _, err := e.adb.Screenshot(imagePath); err != nil {
				e.logger.Error(prefix+"Failed to capture screenshot", slog.Any("error", err))
				return false
			}

			if len(step.Analyze) > 0 {
				newState, err := e.analyzer.AnalyzeAndUpdateState(imagePath, gamer, step.Analyze)
				if err != nil {
					e.logger.Error(prefix+"Analyze failed", slog.Any("error", err))
				} else {
					*gamer = *newState
					e.logger.Info(prefix + "Analyze completed and state updated")
				}
			}
		}
	}

	// Wait
	if step.Wait > 0 {
		e.logger.Info(prefix+"Wait", slog.Duration("duration", step.Wait))
		select {
		case <-time.After(step.Wait):
		case <-ctx.Done():
			e.logger.Warn(prefix+"Wait interrupted by context cancel", slog.Duration("wait", step.Wait))
			return true
		}
	}

	if step.If != nil {
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
			e.logger.Info(prefix + "Condition met: executing THEN")
			for _, s := range step.If.Then {
				stopped := e.runStep(ctx, s, indent+1, gamer)
				if stopped {
					return true
				}
			}
		} else if len(step.If.Else) > 0 {
			e.logger.Info(prefix + "Condition NOT met: executing ELSE")
			for _, s := range step.If.Else {
				stopped := e.runStep(ctx, s, indent+1, gamer)
				if stopped {
					return true
				}
			}
		}
	}

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

		err = e.adb.Swipe(x, y, x, y, step.Wait) // ⬅ swipe на то же место с временем
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
