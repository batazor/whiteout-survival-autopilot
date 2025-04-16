package syncer

import (
	"context"
	"log/slog"
	"sync"

	"github.com/batazor/whiteout-survival-autopilot/internal/century"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

// RefreshAllPlayersFromCentury загружает данные всех игроков через Century API и сохраняет их в state.yaml
func RefreshAllPlayersFromCentury(
	ctx context.Context,
	gamers []*domain.Gamer,
	repo repository.StateRepository,
	logger *slog.Logger,
) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var updatedGamers []domain.Gamer

	for _, g := range gamers {
		gamer := g
		wg.Add(1)

		go func() {
			defer wg.Done()

			info, err := century.FetchPlayerInfo(gamer.ID)
			if err != nil {
				logger.Warn("⚠️ Не удалось получить данные игрока из Century", slog.Int("id", gamer.ID), slog.Any("err", err))
				return
			}

			// Обновляем данные
			gamer.Nickname = info.Data.Nickname
			gamer.State = info.Data.KID
			gamer.Avatar = info.Data.AvatarImage
			gamer.Buildings.Furnace.Level = info.Data.StoveLevel

			mu.Lock()
			updatedGamers = append(updatedGamers, *gamer)
			mu.Unlock()

			logger.Info("📥 Игрок обновлён из Century", slog.String("nickname", gamer.Nickname), slog.Int("id", gamer.ID))
		}()
	}

	wg.Wait()

	// 💾 Сохраняем финальный state.yaml
	finalState := &domain.State{Gamers: updatedGamers}
	if err := repo.SaveState(ctx, finalState); err != nil {
		logger.Error("❌ Не удалось сохранить state.yaml после обновления", slog.Any("error", err))
	} else {
		logger.Info("💾 Финальный state.yaml успешно сохранён")
	}
}
