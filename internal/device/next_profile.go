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

func (d *Device) NextProfile(profileIdx, expectedGamerIdx int) {
	ctx := context.Background()

	d.activeProfileIdx = profileIdx

	profile := d.Profiles[profileIdx]
	expected := &profile.Gamer[expectedGamerIdx]

	d.Logger.Info("🎮 Смена активного игрока",
		slog.String("email", profile.Email),
		slog.String("ожидаемый", expected.Nickname),
	)

	// Устанавливаем колбэк (временно, уточним ниже)
	d.FSM.SetCallback(expected)

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

	// Проверка welcome back
	newCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	resp, _ := vision.WaitForText(newCtx, d.ADB, []string{"Welcome"}, time.Second, image.Rectangle{})
	if resp != nil {
		d.Logger.Info("🟢 Клик по кнопке Welcome Back", slog.String("region", "welcome_back_continue_button"))
		if err := d.ADB.ClickRegion("welcome_back_continue_button", d.areaLookup); err != nil {
			d.Logger.Error("❌ Не удалось кликнуть по welcome_back_continue_button", slog.Any("err", err))
			panic(fmt.Sprintf("ClickRegion(welcome_back_continue_button) failed: %v", err))
		}
	}

	// 📸 Определяем активного игрока после входа
	tmpPath := "screenshots/after_profile_switch.png"
	pIdx, gIdx, err := d.DetectedGamer(ctx, tmpPath)
	if err != nil || pIdx != profileIdx {
		d.Logger.Warn("⚠️ После входа активный профиль не совпадает", slog.Any("detected_profile", pIdx), slog.Any("err", err))
		return
	}

	d.activeGamerIdx = gIdx
	active := &d.Profiles[pIdx].Gamer[gIdx]

	d.Logger.Info("🔎 Игрок после входа", slog.String("nickname", active.Nickname))

	// Устанавливаем колбэк на того, кто реально активен
	d.FSM.SetCallback(active)

	// Если это НЕ тот, кого мы ожидали → переключаемся
	if active.ID != expected.ID {
		d.Logger.Warn("🛑 Автоматически выбран не тот игрок — делаем переключение",
			slog.String("ожидался", expected.Nickname),
			slog.String("получен", active.Nickname),
		)
		d.NextGamer(profileIdx, expectedGamerIdx)
	}

	// FSM пересоздать
	d.Logger.Info("🔧 Инициализация FSM после профиля")
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.areaLookup)
}
