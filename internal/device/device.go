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
	Name     string
	Profiles []domain.Profile
	Logger   *slog.Logger
	ADB      adb.DeviceController
	FSM      *fsm.GameFSM
}

func New(name string, profiles []domain.Profile, log *slog.Logger, lookup *config.AreaLookup) (*Device, error) {
	log.Info("🔧 Инициализация ADB-контроллера")
	controller, err := adb.NewController(log)
	if err != nil {
		log.Error("❌ Не удалось создать ADB-контроллер", slog.Any("error", err))
		return nil, err
	}

	log.Info("🔧 Инициализация FSM")
	stateFSM := fsm.NewGameFSM(log, controller, lookup)

	return &Device{
		Name:     name,
		Profiles: profiles,
		Logger:   log,
		ADB:      controller,
		FSM:      stateFSM,
	}, nil
}

func (d *Device) Start(ctx context.Context) {
	d.Logger.Info("🚀 Старт девайса")

	// Периодическая смена аккаунта
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				d.Logger.Info("🛑 FSM-цикл остановлен")
				return
			case <-ticker.C:
				d.Logger.Info("🔄 FSM смена Google-аккаунта")
				d.FSM.ForceTo(fsm.StateChiefProfile)
				d.FSM.ForceTo(fsm.StateChiefProfileSetting)
				d.FSM.ForceTo(fsm.StateChiefProfileAccount)
				d.FSM.ForceTo(fsm.StateChiefProfileAccountChangeAccount)
				d.FSM.ForceTo(fsm.StateChiefProfileAccountChangeGoogle)
				d.FSM.ForceTo(fsm.StateChiefProfileAccountChangeGoogleConfirm)
				d.FSM.ForceTo(fsm.StateMainCity)
			}
		}
	}()

	// Демонстрация активных игроков
	for {
		for _, profile := range d.Profiles {
			for _, gamer := range profile.Gamer {
				select {
				case <-ctx.Done():
					d.Logger.Info("🛑 Игровой цикл остановлен")
					return
				default:
					d.Logger.Info("▶️ Активный игрок",
						slog.String("email", profile.Email),
						slog.String("nickname", gamer.Nickname),
						slog.Int("id", gamer.ID),
					)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}
