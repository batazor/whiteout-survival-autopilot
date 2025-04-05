package device

import (
	"context"
	"image"
	"log/slog"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func (d *Device) handleEntryScreens(ctx context.Context) error {
	d.Logger.Info("🔎 Проверка экранов входа (welcome / реклама)")

	allKeywords := []string{
		"Welcome",     // welcome back
		"Alliance",    // альянс
		"natalia",     // реклама
		"Exploration", // главный экран — выходим
	}

	timeout := 20 * time.Second
	start := time.Now()

	// Свайп для закрытия стартовых экранов и на всякий случай
	defer d.swipeToDismiss()

	for time.Since(start) < timeout {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			result, _ := vision.WaitForText(ctx, d.ADB, allKeywords, time.Second, image.Rectangle{})
			if result != nil {
				text := strings.ToLower(strings.TrimSpace(result.Text))
				d.Logger.Info("🧠 Обнаружен текст: " + text)

				switch {
				case strings.Contains(text, "exploration"), strings.Contains(text, "alliance"):
					d.Logger.Info("✅ Обнаружен основной экран (Exploration) — выходим из handleEntryScreens")
					return nil

				case strings.Contains(text, "welcome"), strings.Contains(text, "natalia"):
					d.Logger.Info("🌀 Обнаружен pop-up ('%s') — выполняем свайп для закрытия", text)
					err := d.ADB.ClickRegion("ad_banner_close", d.areaLookup)
					if err != nil {
						d.Logger.Error("❌ Не удалось закрыть pop-up", slog.Any("err", err))
						return err
					}

					time.Sleep(100 * time.Millisecond)
				}
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
