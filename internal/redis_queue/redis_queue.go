package redis_queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

const (
	// screenBoost – сколько «виртуальных» очков получает UseCase, чей Node
	// совпадает с текущим экраном. Больше — агрессивнее обрабатываем задачи
	// текущего экрана прежде, чем переходить на другие.
	screenBoost = 5

	// scanWindow – сколько верхних элементов ZSET мы смотрим, прежде чем выбрать
	// лучший. 10–50 достаточно, чтобы почти всегда «видеть» все UC текущего
	// экрана и при этом не перегружать Redis.
	scanWindow = 20
)

// Queue хранит задачи конкретного бота (или геймера) в
// приоритетной очереди Redis (sorted‑set).
//   - Чем выше uc.Priority (0–100), тем ниже score, тем левее элемент в ZSET.
//   - PopBest дополнительно сдвигает score «своего» экрана на screenBoost.
//
// Формат value = json.Marshal(domain.UseCase).
// Score вычисляется при Push: score = 100 - uc.Priority.
// -----------------------------------------------------------------------------
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

// -----------------------------------------------------------------------------
// PopBest – извлекает UC c учётом "бустa" текущего экрана.
// -----------------------------------------------------------------------------

// currentNode – FSM‑узел, на котором сейчас находится UI (например, "alliance").
//
// Алгоритм:
//  1. Берём первые scanWindow элементов из ZSET (они уже грубо отсортированы
//     по приоритету).
//  2. Для каждого декодируем UseCase, считаем adjustedScore = score - screenBoost,
//     если uc.Node == currentNode.
//  3. Выбираем минимальный adjustedScore.
//  4. Удаляем этот элемент из ZSET (ZREM) и возвращаем.
func (q *Queue) PopBest(ctx context.Context, currentNode string) (*domain.UseCase, error) {
	items, err := q.rdb.ZRangeWithScores(ctx, q.key(), 0, scanWindow-1).Result()
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil // очередь пуста – не считаем это ошибкой
	}

	bestIdx, bestScore := -1, 1e9
	var bestUC domain.UseCase

	for i, it := range items {
		raw, ok := it.Member.(string)
		if !ok {
			continue // пропускаем повреждённый элемент
		}
		var uc domain.UseCase
		if err := json.Unmarshal([]byte(raw), &uc); err != nil {
			continue // тоже повреждённый – игнорируем
		}

		score := it.Score
		if uc.Node == currentNode {
			score -= screenBoost
		}

		if score < bestScore {
			bestIdx, bestScore, bestUC = i, score, uc
		}
	}

	if bestIdx == -1 {
		// Все элементы были битые – очищаем окно на всякий случай.
		_ = q.rdb.ZRem(ctx, q.key(), extractMembers(items)...)
		return nil, fmt.Errorf("no decodable use cases in queue window")
	}

	// Удаляем выбранный элемент по оригинальному Member.
	if err := q.rdb.ZRem(ctx, q.key(), items[bestIdx].Member).Err(); err != nil {
		return nil, err
	}

	return &bestUC, nil
}

// extractMembers собирает значения Member из массива redis.Z для служебного
// удаления битых записей.
func extractMembers(zs []redis.Z) []interface{} {
	out := make([]interface{}, 0, len(zs))
	for _, z := range zs {
		out = append(out, z.Member)
	}
	return out
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
