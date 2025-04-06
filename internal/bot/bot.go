package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
)

type Bot struct {
	Gamer  *domain.Gamer
	Device *device.Device
	Queue  *redis_queue.Queue
	Logger *slog.Logger
}

func NewBot(dev *device.Device, gamer *domain.Gamer, rdb *redis.Client, log *slog.Logger) *Bot {
	return &Bot{
		Gamer:  gamer,
		Device: dev,
		Queue:  redis_queue.NewGamerQueue(rdb, gamer.ID),
		Logger: log,
	}
}

func (b *Bot) Play(ctx context.Context) {
	for {
		uc, err := b.Queue.Pop(ctx)
		if err != nil {
			b.Logger.Warn("⚠️ Не удалось получить use‑case", "err", err)
			continue
		}

		// очередь пуста → выходим, чтобы переключиться на другого игрока
		if uc == nil {
			b.Logger.Info("📭 Очередь пуста — завершаю работу бота")
			return
		}

		b.Logger.Info("🚀 Выполняю use‑case", "name", uc.Name, "priority", uc.Priority)

		// переходим на стартовый экран юзкейса
		b.Device.FSM.ForceTo(uc.Node)
		b.Device.Executor.ExecuteUseCase(ctx, uc, b.Gamer)

		time.Sleep(2 * time.Second) // лёгкая пауза между задачами
	}

	// 🔁 Возвращаемся в главный экран
	b.Device.FSM.ForceTo(fsm.StateMainCity)

	b.Logger.Info("⏭️ Очередь завершена. Готов к переключению.")
}
