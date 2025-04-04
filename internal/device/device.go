package device

import (
	"context"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

type Device struct {
	Name       string
	Profiles   domain.Profiles
	Logger     *slog.Logger
	ADB        adb.DeviceController
	FSM        *fsm.GameFSM
	areaLookup *config.AreaLookup

	activeProfileIdx int
	activeGamerIdx   int
}

func New(name string, profiles domain.Profiles, log *slog.Logger, areaPath string) (*Device, error) {
	log.Info("🔧 Инициализация ADB-контроллера")
	controller, err := adb.NewController(log, name)
	if err != nil {
		log.Error("❌ Не удалось создать ADB-контроллер", slog.Any("error", err))
		return nil, err
	}

	areaLookup, err := config.LoadAreaReferences(areaPath)
	if err != nil {
		log.Error("❌ Ошибка загрузки area.json:", "error", err)
		return nil, err
	}

	return &Device{
		Name:       name,
		Profiles:   profiles,
		Logger:     log,
		ADB:        controller,
		FSM:        fsm.NewGame(log, controller, areaLookup),
		areaLookup: areaLookup,
	}, nil
}

func (d *Device) Start(ctx context.Context) {
	d.Logger.Info("🚀 Старт девайса")

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
