package device

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

type Device struct {
	Name       string
	Profiles   []domain.Profile
	Logger     *slog.Logger
	ADB        adb.DeviceController
	FSM        *fsm.GameFSM
	areaLookup *config.AreaLookup

	activeProfileIdx int
	activeGamerIdx   int
}

func New(name string, profiles []domain.Profile, log *slog.Logger) (*Device, error) {
	log.Info("🔧 Инициализация ADB-контроллера")
	controller, err := adb.NewController(log, name)
	if err != nil {
		log.Error("❌ Не удалось создать ADB-контроллер", slog.Any("error", err))
		return nil, err
	}

	areaLookup, err := config.LoadAreaReferences("./references/area.json")
	if err != nil {
		log.Error("❌ Ошибка загрузки area.json: %v", err)
		return nil, err
	}

	log.Info("🔧 Инициализация FSM")
	stateFSM := fsm.NewGame(log, controller, areaLookup)

	return &Device{
		Name:       name,
		Profiles:   profiles,
		Logger:     log,
		ADB:        controller,
		FSM:        stateFSM,
		areaLookup: areaLookup,
	}, nil
}

func (d *Device) Start(ctx context.Context) {
	d.Logger.Info("🚀 Старт девайса")

	// Демонстрация активных игроков
	for {
		for pIdx, profile := range d.Profiles {
			for gIdx := range profile.Gamer {
				select {
				case <-ctx.Done():
					return
				default:
					d.SetActiveGamer(pIdx, gIdx)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

func (d *Device) SetActiveGamer(profileIdx, gamerIdx int) {
	d.activeProfileIdx = profileIdx
	d.activeGamerIdx = gamerIdx

	profile := d.Profiles[profileIdx]
	gamer := &profile.Gamer[gamerIdx]

	d.Logger.Info("🎮 Смена активного игрока",
		slog.String("email", profile.Email),
		slog.String("nickname", gamer.Nickname),
		slog.Int("id", gamer.ID),
	)

	// Устанавливаем колбэк для FSM
	d.FSM.SetCallback(gamer)

	// 🔁 Навигация: переходим к экрану выбора аккаунта Google
	d.Logger.Info("➡️ Переход в экран выбора аккаунта")
	d.FSM.ForceTo(fsm.StateChiefProfileAccountChangeGoogle)

	d.Logger.Info("🟢 Клик по email аккаунту", slog.String("region", "email:gamer1"))
	if err := d.ADB.ClickRegion("email:gamer1", d.areaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по email аккаунту", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(email:gamer1) failed: %v", err))
	}

	time.Sleep(5 * time.Second)

	d.Logger.Info("🟢 Клик по кнопке продолжения Google", slog.String("region", "to_google_continue"))
	if err := d.ADB.ClickRegion("to_google_continue", d.areaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по to_google_continue", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(to_google_continue) failed: %v", err))
	}

	time.Sleep(20 * time.Second)

	// TODO: check ads

	d.Logger.Info("🟢 Клик по кнопке Welcome Back", slog.String("region", "welcome_back_continue_button"))
	if err := d.ADB.ClickRegion("welcome_back_continue_button", d.areaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по welcome_back_continue_button", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(welcome_back_continue_button) failed: %v", err))
	}

	d.Logger.Info("✅ Вход выполнен, переход в Main City")
	d.FSM.ForceTo(fsm.StateMainCity)
}
