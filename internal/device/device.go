package device

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
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

func (d *Device) NextGamer(profileIdx, gamerIdx int) {
	ctx := context.Background()

	d.activeProfileIdx = profileIdx
	d.activeGamerIdx = gamerIdx

	profile := d.Profiles[profileIdx]
	gamer := &profile.Gamer[gamerIdx]

	d.Logger.Info("🎮 Переключение на другого игрока в текущем профиле",
		slog.String("email", profile.Email),
		slog.String("nickname", gamer.Nickname),
		slog.Int("id", gamer.ID),
	)

	// Устанавливаем нового игрока в FSM
	d.FSM.SetCallback(gamer)

	// 🔁 Навигация: переходим к экрану выбора аккаунта Google
	d.Logger.Info("➡️ Переход в экран выбора игрока")
	d.FSM.ForceTo(fsm.StateChiefCharacters)

	// ждем nickname
	gamerZones, _ := vision.WaitForText(ctx, d.ADB, []string{gamer.Nickname}, time.Second, image.Rectangle{})

	d.Logger.Info("🟢 Клик по nickname игрока", slog.String("text", gamerZones.Text))
	if err := d.ADB.ClickOCRResult(gamerZones); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по nickname аккаунту", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(nickname:%s) failed: %v", gamer.Nickname, err))
	}

	time.Sleep(2 * time.Second)

	d.Logger.Info("🟢 Клик по кнопке подтверждения", slog.String("region", "character_change_confirm"))
	if err := d.ADB.ClickRegion("character_change_confirm", d.areaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по character_change_confirm", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(character_change_confirm) failed: %v", err))
	}

	// Проверка на страницу - добро пожаловать
	newCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	resp, _ := vision.WaitForText(newCtx, d.ADB, []string{"Welcome"}, time.Second, image.Rectangle{})

	if resp != nil {
		d.Logger.Info("🟢 Клик по кнопке Welcome Back", slog.String("region", "welcome_back_continue_button"))
		if err := d.ADB.ClickRegion("welcome_back_continue_button", d.areaLookup); err != nil {
			d.Logger.Error("❌ Не удалось кликнуть по welcome_back_continue_button", slog.Any("err", err))
			panic(fmt.Sprintf("ClickRegion(welcome_back_continue_button) failed: %v", err))
		}
	}

	d.Logger.Info("✅ Вход выполнен, переход в Main City")
	d.Logger.Info("🔧 Инициализация FSM")
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.areaLookup)
}

func (d *Device) NextProfile(profileIdx, gamerIdx int) {
	ctx := context.Background()

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

	// ждем email
	emailZones, _ := vision.WaitForText(ctx, d.ADB, []string{profile.Email}, time.Second, image.Rectangle{})

	d.Logger.Info("🟢 Клик по email аккаунту", slog.String("text", emailZones.Text))
	if err := d.ADB.ClickOCRResult(emailZones); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по email аккаунту", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(email:gamer1) failed: %v", err))
	}

	time.Sleep(5 * time.Second)

	d.Logger.Info("🟢 Клик по кнопке продолжения Google", slog.String("region", "to_google_continue"))
	if err := d.ADB.ClickRegion("to_google_continue", d.areaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по to_google_continue", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(to_google_continue) failed: %v", err))
	}

	// Проверка на страницу - добро пожаловать
	newCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	resp, _ := vision.WaitForText(newCtx, d.ADB, []string{"Welcome"}, time.Second, image.Rectangle{})

	if resp != nil {
		d.Logger.Info("🟢 Клик по кнопке Welcome Back", slog.String("region", "welcome_back_continue_button"))
		if err := d.ADB.ClickRegion("welcome_back_continue_button", d.areaLookup); err != nil {
			d.Logger.Error("❌ Не удалось кликнуть по welcome_back_continue_button", slog.Any("err", err))
			panic(fmt.Sprintf("ClickRegion(welcome_back_continue_button) failed: %v", err))
		}
	}

	d.Logger.Info("✅ Вход выполнен, переход в Main City")
	d.Logger.Info("🔧 Инициализация FSM")
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.areaLookup)
}
