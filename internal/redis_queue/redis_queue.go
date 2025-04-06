package redis_queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type Queue struct {
	rdb   *redis.Client
	botID string
}

func (q *Queue) key() string {
	return fmt.Sprintf("bot:queue:%s", q.botID)
}

func NewGamerQueue(rdb *redis.Client, gamerID int) *Queue {
	return &Queue{rdb: rdb, botID: fmt.Sprintf("gamer:%d", gamerID)}
}

// Push добавляет UseCase в приоритетную очередь Redis.
// Чем выше uc.Priority (0–100), тем выше приоритет задачи (меньше score).
func (q *Queue) Push(ctx context.Context, uc *domain.UseCase) error {
	data, err := json.Marshal(uc)
	if err != nil {
		return err
	}

	score := float64(100 - uc.Priority) // Чем выше приоритет, тем ниже score
	return q.rdb.ZAdd(ctx, q.key(), redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

// Pop извлекает самый приоритетный UseCase из очереди Redis.
func (q *Queue) Pop(ctx context.Context) (*domain.UseCase, error) {
	items, err := q.rdb.ZPopMin(ctx, q.key(), 1).Result()
	if err != nil || len(items) == 0 {
		return nil, err
	}

	var uc domain.UseCase
	if err := json.Unmarshal([]byte(items[0].Member.(string)), &uc); err != nil {
		return nil, err
	}

	return &uc, nil
}

// Peek возвращает самый приоритетный UseCase без удаления (полезно для анализа)
func (q *Queue) Peek(ctx context.Context) (*domain.UseCase, error) {
	items, err := q.rdb.ZRangeWithScores(ctx, q.key(), 0, 0).Result()
	if err != nil || len(items) == 0 {
		return nil, err
	}

	var uc domain.UseCase
	if err := json.Unmarshal([]byte(items[0].Member.(string)), &uc); err != nil {
		return nil, err
	}

	return &uc, nil
}

// Len возвращает количество задач в очереди
func (q *Queue) Len(ctx context.Context) (int64, error) {
	return q.rdb.ZCard(ctx, q.key()).Result()
}
