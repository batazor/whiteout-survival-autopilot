package redis_queue

import (
	"context"
	"fmt"
	"time"
)

func (q *Queue) SetLastExecuted(ctx context.Context, botID int, usecaseName string, ttl time.Duration) error {
	key := fmt.Sprintf("bot:last_executed:%d:%s", botID, usecaseName)
	return q.rdb.Set(ctx, key, time.Now().Unix(), ttl).Err()
}

func (q *Queue) ShouldSkip(ctx context.Context, botID int, usecaseName string) (bool, error) {
	key := fmt.Sprintf("bot:last_executed:%d:%s", botID, usecaseName)

	exists, err := q.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}
