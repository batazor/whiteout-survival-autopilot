package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
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
	exec := executor.NewUseCaseExecutor(
		log,
		config.NewTriggerEvaluator(),
		analyzer.NewAnalyzer(dev.AreaLookup, log),
		dev.ADB,
		dev.AreaLookup,
		gamer.Nickname,
	)

	return &Bot{
		Gamer:    gamer,
		Email:    email,
		Device:   dev,
		Queue:    redis_queue.NewGamerQueue(rdb, gamer.ID),
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
		uc, err := b.Queue.Pop(ctx)
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
		b.Device.FSM.ForceTo(uc.Node)

		// 📸 Анализ состояния перед trigger'ом
		b.updateStateFromScreen(ctx, uc.Node, "out/bot_"+b.Gamer.Nickname+"_before_trigger.png")

		b.executor.ExecuteUseCase(ctx, uc, b.Gamer, b.Queue)

		// Время для отрисовки экрана
		time.Sleep(1 * time.Second)
	}

	// Время для отрисовки экрана
	time.Sleep(2 * time.Second)

	// 🔁 Возвращаемся в главный экран
	b.Device.FSM.ForceTo(fsm.StateMainCity)

	// Время для отрисовки экрана
	time.Sleep(1 * time.Second)

	b.logger.Info("⏭️ Очередь завершена. Готов к переключению.")
}
