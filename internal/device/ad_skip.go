package device

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/imagefinder"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func (d *Device) handleEntryScreens(ctx context.Context) error {
	d.Logger.Info("🔎 Проверка экранов входа (welcome / реклама)")

	allKeywords := []string{
		"Welcome",
		"Alliance",
		"natalia",
		"Exploration",
		"Hero Gear",
		"General Speedup",
		"Construction Speedup",
	}

	timeout := 20 * time.Second
	start := time.Now()

	defer d.swipeToDismiss()

	mainScreenDetectedAt := time.Time{}

	for time.Since(start) < timeout {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// проверяем текст на экране
			result, _ := vision.WaitForText(ctx, d.ADB, allKeywords, time.Second, image.Rectangle{})
			if result != nil {
				text := strings.ToLower(strings.TrimSpace(result.Text))
				d.Logger.Info("🧠 Обнаружен текст: " + text)

				switch {
				case strings.Contains(text, "exploration"), strings.Contains(text, "alliance"):
					// 📌 Обнаружен главный экран — запоминаем время
					if mainScreenDetectedAt.IsZero() {
						d.Logger.Info("🔔 Найден главный экран — контрольная пауза")
						mainScreenDetectedAt = time.Now()
					}

				case strings.Contains(text, "welcome"),
					strings.Contains(text, "natalia"),
					strings.Contains(text, "hero gear"),
					strings.Contains(text, "general speedup"),
					strings.Contains(text, "construction speedup"):
					d.Logger.Info(fmt.Sprintf("🌀 Найден pop-up ('%s') — закрываем", text))
					err := d.ADB.ClickRegion("ad_banner_close", d.areaLookup)
					if err != nil {
						d.Logger.Error("❌ Не удалось закрыть pop-up", slog.Any("err", err))
						return err
					}
					time.Sleep(300 * time.Millisecond)
					mainScreenDetectedAt = time.Time{} // сбрасываем, так как появилось новое окно
				}
			}

			// проверяем наличие кнопки confirm, если есть — кликаем
			// проверяем по наличию зеленого в зоне welcome_back_continue_button, потому что OCR плохо работает
			// на этом экране
			isConfirm, err := imagefinder.CheckRegionColor(ctx, d.ADB, d.areaLookup, "welcome_back_continue_button", "green", 0.3, d.Logger)
			if err != nil {
				d.Logger.Error("❌ Ошибка проверки цвета", slog.Any("err", err))
				return err
			}

			if isConfirm {
				d.Logger.Info("🟢 Клик по кнопке продолжения welcome_back_continue_button")
				if err := d.ADB.ClickRegion("welcome_back_continue_button", d.areaLookup); err != nil {
					d.Logger.Error("❌ Не удалось кликнуть по welcome_back_continue_button", slog.Any("err", err))
					return err
				}

				time.Sleep(1 * time.Second)
			}

			// ✅ Если главный экран был замечен и прошло >2.5 секунды — считаем, что всё чисто
			if !mainScreenDetectedAt.IsZero() && time.Since(mainScreenDetectedAt) > 2500*time.Millisecond {
				d.Logger.Info("✅ Подтверждён основной экран — выходим из handleEntryScreens")
				return nil
			}

			time.Sleep(300 * time.Millisecond)
		}
	}

	d.Logger.Warn("⏱ Ничего не найдено — выполняем проактивный свайп")
	return nil
}

func (d *Device) swipeToDismiss() {
	const (
		swipeY = 500
		dx     = 150
		delay  = 100 * time.Millisecond
	)

	_ = d.ADB.Swipe(500, swipeY, 500-dx, swipeY, delay)
	time.Sleep(300 * time.Millisecond)
	_ = d.ADB.Swipe(500-dx, swipeY, 500, swipeY, delay)
}
