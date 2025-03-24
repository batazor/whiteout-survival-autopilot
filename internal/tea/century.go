package teaapp

import (
	"log/slog"

	"github.com/batazor/whiteout-survival-autopilot/internal/century"
)

func (a *App) UpdateCharacterInfoFromCentury() {
	for i := range a.state.Accounts {
		for j := range a.state.Accounts[i].Characters {
			char := &a.state.Accounts[i].Characters[j]

			playerInfo, err := century.FetchPlayerInfo(char.ID)
			if err != nil {
				a.logger.Warn("failed to fetch player info",
					slog.Int64("fid", int64(char.ID)),
					slog.String("nickname", char.Nickname),
					slog.Any("error", err),
				)
				continue
			}

			char.Nickname = playerInfo.Data.Nickname
			char.State = playerInfo.Data.KID
			char.Buildings.Furnace.Level = playerInfo.Data.StoveLevel
			char.Avatar = playerInfo.Data.AvatarImage

			a.logger.Info("updated character info",
				slog.Int64("fid", int64(char.ID)),
				slog.String("nickname", char.Nickname),
				slog.Int("state", char.State),
				slog.Int("furnace_level", char.Buildings.Furnace.Level),
			)
		}
	}
}
