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

	// 🔧 Пересоздаём FSM для нового аккаунта до любых ForceTo/WaitForText
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.areaLookup)

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

	// сбросим FSM
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.areaLookup)

	// Проверка стартовых баннеров
	err := d.handleEntryScreens(ctx)
	if err != nil {
		d.Logger.Error("❌ Не удалось обработать стартовые баннеры", slog.Any("err", err))
		panic(fmt.Sprintf("handleEntryScreens() failed: %v", err))
	}

	// Проверяем, что мы находимся на экране профиля
	active, pIdx, _, err := d.DetectAndSetCurrentGamer(ctx)
	if err != nil || pIdx != profileIdx {
		d.Logger.Warn("⚠️ После входа активный профиль не совпадает", slog.Any("detected_profile", pIdx), slog.Any("err", err))
		return
	}

	// Если это НЕ тот, кого мы ожидали → переключаемся
	if active.ID != expected.ID {
		d.Logger.Warn("🛑 Автоматически выбран не тот игрок — делаем переключение",
			slog.String("ожидался", expected.Nickname),
			slog.String("получен", active.Nickname),
		)
		d.NextGamer(profileIdx, expectedGamerIdx)
	}

	// Устанавливаем колбэк (настоящий)
	d.FSM.SetCallback(active)

	// Успешно переключились на новый профиль
	d.Logger.Info("✅ Успешно переключились на новый профиль", "nickname", active.Nickname)
}
