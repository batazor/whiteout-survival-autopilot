package device

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
)

func (d *Device) NextProfile(profileIdx, expectedGamerIdx int) {
	// 🕒 Ждём, чтобы не было конфликта с другими процессами
	time.Sleep(500 * time.Millisecond)

	ctx := context.Background()

	profile := d.Profiles[profileIdx]
	expected := &profile.Gamer[expectedGamerIdx]

	d.Logger.Info("🎮 Смена активного игрока",
		slog.String("email", profile.Email),
		slog.String("ожидаемый", expected.Nickname),
	)

	// 🔁 Навигация: переходим к экрану выбора аккаунта Google
	d.Logger.Info("➡️ Переход в экран выбора аккаунта")
	d.FSM.ForceTo(state.StateChiefProfileAccountChangeGoogle, nil)

	// 🕒 Ждём, чтобы не было конфликта с другими процессами
	time.Sleep(2 * time.Second)

	// ========== 1️⃣ Делаем единый full-screen OCR ==========
	region, ok := d.AreaLookup.Get("google_profile")
	if !ok {
		d.Logger.Error("❌ Не удалось найти область google_profile")
		panic("AreaLookup(google_profile) failed")
	}

	fullOCR, fullErr := d.OCRClient.FetchOCR("google_profile", []ocrclient.Region{
		{
			X0: region.Zone.Min.X,
			Y0: region.Zone.Min.Y,
			X1: region.Zone.Max.X,
			Y1: region.Zone.Max.Y,
		},
	})
	if fullErr != nil {
		d.Logger.Error("❌ Full OCR failed", slog.Any("error", fullErr))
		panic(fmt.Sprintf("ocrClient.FetchOCR() failed: %v", fullErr))
	}

	// 📦 OCR по email
	var emailZone *domain.OCRResult
	for _, zone := range fullOCR {
		if zone.Text == profile.Email {
			emailZone = &zone
			break
		}
	}

	d.Logger.Info("🟢 Клик по email аккаунту", slog.String("text", emailZone.Text), slog.String("region", emailZone.String()))
	if err := d.ADB.ClickOCRResult(emailZone); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по email аккаунту", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(email:gamer1) failed: %v", err))
	}

	time.Sleep(3 * time.Second)

	googleContinueArea, ok := d.AreaLookup.Get("to_google_continue")
	if !ok {
		d.Logger.Error("❌ Не удалось найти область to_google_continue")
		panic("AreaLookup(to_google_continue) failed")
	}

	// Ждём текст "Continue" через OCR-клиент
	if _, err := d.OCRClient.WaitForText([]string{"Continue"}, time.Second, 500*time.Millisecond, "continue"); err != nil {
		d.Logger.Error("❌ OCRClient WaitForText failed for Continue", slog.Any("err", err))
		panic(fmt.Sprintf("OCRClient.WaitForText(Continue) failed: %v", err))
	}

	d.Logger.Info("🟢 Клик по кнопке продолжения Google", slog.String("region", "to_google_continue"))

	if err := d.ADB.Click(googleContinueArea.Zone); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по to_google_continue", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(to_google_continue) failed: %v", err))
	}

	// ♻️ сброс FSM после входа
	d.activeProfileIdx = profileIdx
	d.activeGamerIdx = expectedGamerIdx
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.AreaLookup, d.triggerEvaluator, d.ActiveGamer(), d.OCRClient)

	// Проверка стартовых баннеров
	err := d.handleEntryScreens(ctx)
	if err != nil {
		d.Logger.Error("❌ Не удалось обработать стартовые баннеры", slog.Any("err", err))
		panic(fmt.Sprintf("handleEntryScreens() failed: %v", err))
	}

	// 🔍 Проверяем, что активный профиль — тот, что ожидали
	active, pIdx, _, err := d.DetectAndSetCurrentGamer(ctx)
	if err != nil || pIdx != profileIdx {
		d.Logger.Warn("⚠️ После входа активный профиль не совпадает", slog.Any("detected_profile", pIdx), slog.Any("err", err))
		return
	}

	// 🧾 Если игрок не тот — переключаемся вручную
	if active.ID != expected.ID {
		d.Logger.Warn("🛑 Автоматически выбран не тот игрок — делаем переключение",
			slog.String("ожидался", expected.Nickname),
			slog.String("получен", active.Nickname),
		)
		d.NextGamer(profileIdx, expectedGamerIdx)
	}

	// ✅ Устанавливаем колбэк
	d.FSM.SetCallback(active)

	d.Logger.Info("✅ Успешно переключились на новый профиль", "nickname", active.Nickname)
}

func (d *Device) ActiveGamer() *domain.Gamer {
	if d.activeProfileIdx >= 0 && d.activeProfileIdx < len(d.Profiles) {
		profile := d.Profiles[d.activeProfileIdx]
		if d.activeGamerIdx >= 0 && d.activeGamerIdx < len(profile.Gamer) {
			return &profile.Gamer[d.activeGamerIdx]
		}
	}
	return nil
}
