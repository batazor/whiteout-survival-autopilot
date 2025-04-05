package device

import (
	"context"
	"time"
)

func (d *Device) Start(ctx context.Context) {
	d.Logger.Info("🚀 Старт девайса")

	// ✅ Проверяем кто уже активен
	if _, pIdx, gIdx, err := d.DetectAndSetCurrentGamer(ctx); err == nil {
		// Пропускаем текущего и начинаем с next
		if gIdx+1 < len(d.Profiles[pIdx].Gamer) {
			gIdx++
		} else {
			pIdx++
			gIdx = 0
		}

		d.activeProfileIdx = pIdx
		d.activeGamerIdx = gIdx
	}

	for {
		for pIdx, profile := range d.Profiles {
			for gIdx := range profile.Gamer {
				select {
				case <-ctx.Done():
					d.Logger.Info("🛑 Остановка девайса по контексту")
					return
				default:
					if gIdx == 0 {
						d.Logger.Info("🔄 Смена профиля и переход к первому игроку",
							"profile_index", pIdx,
							"gamer_index", gIdx,
							"nickname", profile.Gamer[gIdx].Nickname,
						)
						d.NextProfile(pIdx, gIdx)
					} else {
						d.Logger.Info("👤 Переход к следующему игроку того же профиля",
							"profile_index", pIdx,
							"gamer_index", gIdx,
							"nickname", profile.Gamer[gIdx].Nickname,
						)
						d.NextGamer(pIdx, gIdx)
					}

					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}
