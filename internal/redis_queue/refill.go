package redis_queue

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
)

func StartGlobalUsecaseRefiller(
	ctx context.Context,
	cfg *domain.Config,
	usecasePath string,
	rdb *redis.Client,
	log *logger.TracedLogger,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	usecaseLoader := config.NewUseCaseLoader(usecasePath)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info(ctx, "🛑 Остановка глобального рефиллера задач")
				return
			case <-ticker.C:
				log.Info(ctx, "🔄 Запуск глобального refill задач")

				usecases, err := usecaseLoader.LoadAll(ctx)
				if err != nil {
					log.Error(ctx, "❌ Не удалось загрузить usecases", slog.Any("err", err))
					continue
				}

				allGamers := cfg.AllGamers()

				for _, gamer := range allGamers {
					queue := NewGamerQueue(rdb, gamer.ID)

					for _, uc := range usecases {
						shouldSkip, err := queue.ShouldSkip(ctx, gamer.ID, uc.Name)
						if err != nil {
							log.Warn(ctx, "⚠️ Ошибка проверки TTL", slog.Any("botID", gamer.ID), slog.Any("usecase", uc.Name), slog.Any("err", err))
							continue
						}

						if shouldSkip {
							continue
						}

						if err := queue.Push(ctx, uc); err != nil {
							log.Error(ctx, "❌ Не удалось добавить usecase", slog.Any("botID", gamer.ID), slog.Any("usecase", uc.Name), slog.Any("err", err))
						} else {
							log.Info(ctx, "✅ Usecase добавлен", slog.Any("usecase", uc.Name), slog.Any("botID", gamer.ID))
						}
					}
				}
			}
		}
	}()
}
