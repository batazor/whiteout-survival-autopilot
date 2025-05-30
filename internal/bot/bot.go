package bot

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

type Bot struct {
	Gamer    *domain.Gamer
	Email    string
	Device   *device.Device
	Queue    *redis_queue.Queue
	logger   *slog.Logger
	Rules    config.ScreenAnalyzeRules
	Repo     repository.StateRepository
	executor executor.UseCaseExecutor
}

func NewBot(dev *device.Device, gamer *domain.Gamer, email string, rdb *redis.Client, rules config.ScreenAnalyzeRules, log *slog.Logger, repo repository.StateRepository) *Bot {
	queue := redis_queue.NewGamerQueue(rdb, gamer.ID)

	exec := executor.NewUseCaseExecutor(
		log,
		config.NewTriggerEvaluator(),
		analyzer.NewAnalyzer(dev.AreaLookup, log, dev.OCRClient),
		dev.ADB,
		dev.AreaLookup,
		gamer.Nickname,
		queue,
	)

	return &Bot{
		Gamer:    gamer,
		Email:    email,
		Device:   dev,
		Queue:    queue,
		logger:   log,
		Rules:    rules,
		Repo:     repo,
		executor: exec,
	}
}

func (b *Bot) Play(ctx context.Context) {
	// 📸 Анализ состояния на главном экране
	b.updateStateFromScreen(ctx, "main_city", "out/bot_"+b.Gamer.Nickname+"_start_main_city.png")

	for {
		select {
		case <-ctx.Done():
			b.logger.Warn("🛑 Контекст отменён — завершаю работу бота")
			return
		default:
		}

		// получаем use‑case из очереди
		uc, err := b.Queue.PopBest(ctx, b.Gamer.ScreenState.CurrentState)
		if err != nil {
			b.logger.Warn("⚠️ Не удалось получить use‑case", "err", err)
			continue
		}

		// очередь пуста → выходим, чтобы переключиться на другого игрока
		if uc == nil {
			b.logger.Info("📭 Очередь пуста — завершаю работу бота")
			break
		}

		// 🕒 Проверка TTL (пропускаем usecase, если не истёк)
		shouldSkip, err := b.Queue.ShouldSkip(ctx, b.Gamer.ID, uc.Name)
		if err != nil {
			b.logger.Error("❌ Не удалось проверить TTL usecase", slog.Any("err", err))
			continue
		}
		if shouldSkip {
			b.logger.Info("⏭️ UseCase пропущен по TTL", slog.String("name", uc.Name))
			continue
		}

		b.logger.Info("🚀 Выполняю use‑case", "name", uc.Name, "priority", uc.Priority)

		// переходим на стартовый экран юзкейса
		switchedScreen := false
		if b.Gamer.ScreenState.CurrentState != uc.Node {
			b.logger.Info("🔁 Переключаюсь на экран usecase", slog.String("name", uc.Name), slog.String("screen", uc.Node))
			errForceTo := b.Device.FSM.ForceTo(uc.Node, b.updateStateFromScreen)
			if errForceTo != nil {
				if errors.Is(errForceTo, fsm.EventNotActive) {
					b.logger.Info("⏭️ UseCase пропущен, так как событие не активно", slog.String("name", uc.Name))

					// Устанавливает TTL для usecase в очереди
					errSetLastExecuted := b.Queue.SetLastExecuted(ctx, b.Gamer.ID, uc.Name, uc.TTL)
					if errSetLastExecuted != nil {
						b.logger.Error("❌ Не удалось установить TTL usecase", slog.Any("err", err))
					}

					continue
				}

				b.logger.Error("❌ Не удалось переключиться на экран usecase", slog.Any("err", errForceTo))
			} else {
				switchedScreen = true
			}
		} else {
			b.logger.Info("🔁 Находится на экране usecase", slog.String("name", uc.Name), slog.String("screen", uc.Node))
		}

		// Вызываем updateStateFromScreen только если FSM не делал этого в ForceTo, или если не было перехода
		if !switchedScreen {
			b.updateStateFromScreen(ctx, uc.Node, "out/bot_"+b.Gamer.Nickname+"_before_trigger.png")
		}

		b.executor.ExecuteUseCase(ctx, uc, b.Gamer, b.Queue)

		// Время для отрисовки экрана
		time.Sleep(1 * time.Second)
	}

	// Время для отрисовки экрана
	time.Sleep(2 * time.Second)

	// 🔁 Возвращаемся в главный экран
	b.Device.FSM.ForceTo(state.StateMainCity, nil)

	// Время для отрисовки экрана
	time.Sleep(2 * time.Second)

	b.logger.Info("⏭️ Очередь завершена. Готов к переключению.")
}
