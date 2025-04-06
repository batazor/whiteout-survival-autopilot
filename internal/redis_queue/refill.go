package redis_queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func StartGlobalUsecaseRefiller(
	ctx context.Context,
	cfg *domain.Config,
	usecasePath string,
	rdb *redis.Client,
	log *slog.Logger,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	usecaseLoader := config.NewUseCaseLoader(usecasePath)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("🛑 Остановка глобального рефиллера задач")
				return
			case <-ticker.C:
				log.Info("🔄 Запуск глобального refill задач")

				usecases, err := usecaseLoader.LoadAll(ctx)
				if err != nil {
					log.Error("❌ Не удалось загрузить usecases", "err", err)
					continue
				}

				allGamers := cfg.AllGamers()

				for _, gamer := range allGamers {
					queue := NewGamerQueue(rdb, gamer.ID)
					botID := fmt.Sprintf("%d", gamer.ID)

					for _, uc := range usecases {
						shouldSkip, err := queue.ShouldSkip(ctx, botID, uc.Name)
						if err != nil {
							log.Warn("⚠️ Ошибка проверки TTL", "botID", botID, "usecase", uc.Name, "err", err)
							continue
						}

						if shouldSkip {
							continue
						}

						if err := queue.Push(ctx, uc); err != nil {
							log.Error("❌ Не удалось добавить usecase", "usecase", uc.Name, "botID", botID, "err", err)
						} else {
							log.Info("✅ Usecase добавлен", "usecase", uc.Name, "botID", botID)
						}
					}
				}
			}
		}
	}()
}
