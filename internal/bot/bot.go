package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

type Bot struct {
	Gamer  *domain.Gamer
	Device *device.Device
	Queue  *redis_queue.Queue
	Logger *slog.Logger
	Rules  config.ScreenAnalyzeRules
	Repo   repository.StateRepository
}

func NewBot(dev *device.Device, gamer *domain.Gamer, rdb *redis.Client, rules config.ScreenAnalyzeRules, log *slog.Logger, repo repository.StateRepository) *Bot {
	return &Bot{
		Gamer:  gamer,
		Device: dev,
		Queue:  redis_queue.NewGamerQueue(rdb, gamer.ID),
		Logger: log,
		Rules:  rules,
		Repo:   repo,
	}
}

func (b *Bot) Play(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			b.Logger.Warn("🛑 Контекст отменён — завершаю работу бота")
			return
		default:
		}

		// получаем use‑case из очереди
		uc, err := b.Queue.Pop(ctx)
		if err != nil {
			b.Logger.Warn("⚠️ Не удалось получить use‑case", "err", err)
			continue
		}

		// очередь пуста → выходим, чтобы переключиться на другого игрока
		if uc == nil {
			b.Logger.Info("📭 Очередь пуста — завершаю работу бота")
			break
		}

		b.Logger.Info("🚀 Выполняю use‑case", "name", uc.Name, "priority", uc.Priority)

		// переходим на стартовый экран юзкейса
		b.Device.FSM.ForceTo(uc.Node)

		// обновляем state из скриншота перед проверкой trigger'а
		screenshotPath := "out/bot_" + b.Gamer.Nickname + "_before_trigger.png"
		rulesForScreen := b.Rules[uc.Node]

		if _, screenshotErr := b.Device.ADB.Screenshot(screenshotPath); screenshotErr == nil {
			if newState, analyzeErr := b.Device.Executor.Analyzer().AnalyzeAndUpdateState(screenshotPath, b.Gamer, rulesForScreen); analyzeErr == nil {
				*b.Gamer = *newState
				b.Logger.Info("📥 Состояние обновлено перед выполнением usecase", "screen", uc.Node)

				// 🆕 Сохраняем новое состояние в state.yaml
				if err := b.Repo.SaveGamer(ctx, newState); err != nil {
					b.Logger.Error("❌ Не удалось сохранить state.yaml", slog.Any("error", err))
				} else {
					b.Logger.Info("💾 Состояние игрока сохранено в state.yaml")
				}
			} else {
				b.Logger.Warn("⚠️ Ошибка анализа state", "err", analyzeErr)
			}
		} else {
			b.Logger.Warn("⚠️ Не удалось сделать скриншот перед trigger", "err", screenshotErr)
		}

		b.Device.Executor.ExecuteUseCase(ctx, uc, b.Gamer, b.Queue)

		// Время для отрисовки экрана
		time.Sleep(1 * time.Second)
	}

	// Время для отрисовки экрана
	time.Sleep(2 * time.Second)

	// 🔁 Возвращаемся в главный экран
	b.Device.FSM.ForceTo(fsm.StateMainCity)

	// Время для отрисовки экрана
	time.Sleep(1 * time.Second)

	b.Logger.Info("⏭️ Очередь завершена. Готов к переключению.")
}
