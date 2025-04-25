package redis_queue

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func StartGlobalUsecaseRefiller(
	ctx context.Context,
	cfg *domain.Config,
	usecaseLoader config.UseCaseLoader,
	rdb *redis.Client,
	log *slog.Logger,
) {
	usecases, err := usecaseLoader.LoadAll(ctx)
	if err != nil {
		log.Error("❌ Не удалось загрузить usecases при старте", "err", err)
		return
	}

	s, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		log.Error("❌ Не удалось создать gocron scheduler", "err", err)
		return
	}

	for _, uc := range usecases {
		if uc.Cron == "" {
			continue
		}

		ucCopy := uc // замыкание копии usecase

		task := func() {
			allGamers := cfg.AllGamers()
			for _, gamer := range allGamers {
				queue := NewGamerQueue(rdb, gamer.ID)

				shouldSkip, err := queue.ShouldSkip(ctx, gamer.ID, ucCopy.Name)
				if err != nil {
					log.Warn("⚠️ Ошибка проверки TTL", "botID", gamer.ID, "usecase", ucCopy.Name, "err", err)
					continue
				}
				if shouldSkip {
					continue
				}

				if err := queue.Push(ctx, ucCopy); err != nil {
					log.Error("❌ Не удалось добавить usecase", "usecase", ucCopy.Name, "botID", gamer.ID, "err", err)
				} else {
					log.Info("✅ Usecase добавлен", "usecase", ucCopy.Name, "botID", gamer.ID)
				}
			}
		}

		_, err := s.NewJob(
			gocron.CronJob(uc.Cron, true),
			gocron.NewTask(task),
		)
		if err != nil {
			log.Error("❌ Не удалось создать cron-задачу", "cron", uc.Cron, "usecase", uc.Name, "err", err)
		}
	}

	s.Start()
}
