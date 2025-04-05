package redis_queue

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type RedisQueue struct {
	Rdb *redis.Client
	Key string
}

func (q *RedisQueue) Push(ctx context.Context, uc *domain.UseCase) error {
	data, err := json.Marshal(uc)
	if err != nil {
		return err
	}

	score := float64(100 - uc.Priority) // Чем выше приоритет — тем ниже score
	return q.Rdb.ZAdd(ctx, q.Key, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

func (q *RedisQueue) Pop(ctx context.Context) (*domain.UseCase, error) {
	items, err := q.Rdb.ZPopMin(ctx, q.Key, 1).Result()
	if err != nil || len(items) == 0 {
		return nil, err
	}

	var uc domain.UseCase
	if err := json.Unmarshal([]byte(items[0].Member.(string)), &uc); err != nil {
		return nil, err
	}

	return &uc, nil
}
