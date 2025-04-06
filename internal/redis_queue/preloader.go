package redis_queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func PreloadQueues(ctx context.Context, rdb *redis.Client, profiles domain.Profiles, usecasePath string) {
	usecaseLoader := config.NewUseCaseLoader(usecasePath)

	for _, profile := range profiles {
		for _, gamer := range profile.Gamer {
			queue := NewGamerQueue(rdb, gamer.ID) // bot:queue:gamer:<id>
			key := queue.key()

			usecases, err := usecaseLoader.LoadAll(ctx)
			if err != nil {
				fmt.Printf("❌ Ошибка загрузки usecase'ов для gamer:%d: %v\n", gamer.ID, err)
				continue
			}

			for _, uc := range usecases {
				data, _ := json.Marshal(uc)

				// Если такой элемент уже есть в Z‑set — пропускаем
				exists, _ := rdb.ZScore(ctx, key, string(data)).Result()
				if exists != 0 { // элемент найден
					continue
				}

				score := float64(100 - uc.Priority) // чем выше Priority, тем ниже score
				if err := rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: data}).Err(); err != nil {
					fmt.Printf("❌ Не удалось добавить %s в gamer:%d: %v\n", uc.Name, gamer.ID, err)
				} else {
					fmt.Printf("📥 Добавлен usecase %s в gamer:%d\n", uc.Name, gamer.ID)
				}
			}
		}
	}
}
