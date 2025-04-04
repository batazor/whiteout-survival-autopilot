package device

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

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
