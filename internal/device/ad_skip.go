package device

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func (d *Device) handleEntryScreens(ctx context.Context) error {
	d.Logger.Info("🔎 Ждём экранов входа и поп-апов…")

	keywords := []string{
		"Welcome", "Alliance", "natalia", "Exploration", "Hero Gear",
		"General Speedup", "Construction Speedup", "Resource",
		"Mastery Material", "Purchase limit", "Agility",
		"Brothers in Arms", "Event Coming Soon", "Dawn Pack",
		"Unyielding Dawn", "Overview", "Confirm",
	}
	// Приводим все ключи к нижнему регистру
	lowerKW := make([]string, len(keywords))
	for i, kw := range keywords {
		lowerKW[i] = strings.ToLower(kw)
	}

	start := time.Now()
	for {
		// Проверяем общий таймаут 30 секунд
		if time.Since(start) > 30*time.Second {
			d.Logger.Info("⏱️ 30s истекли, выходим без кликов")
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		zones, err := d.OCRClient.WaitForText(lowerKW, 10*time.Second, 1*time.Second, "entry_check")
		if err != nil {
			d.Logger.Error("❌ OCRClient error", slog.Any("err", err))
			return err
		}
		if len(zones) == 0 {
			continue
		}

		// 1) Ищем Confirm
		for _, z := range zones {
			txt := strings.ToLower(strings.TrimSpace(z.Text))
			if vision.FuzzySubstringMatch(txt, "confirm", 1) &&
				z.AvgColor == "white" && z.BgColor == "green" {
				d.Logger.Info("🟢 Кликаем Confirm", slog.String("text", txt))
				if err := d.ADB.ClickRegion("welcome_back_continue_button", d.AreaLookup); err != nil {
					d.Logger.Error("❌ Ошибка клика Confirm", slog.Any("err", err))
					return err
				}
				time.Sleep(time.Second)
				return nil
			}
		}

		// 2) Ищем первый поп-ап
		found := false
		for _, z := range zones {
			txt := strings.ToLower(strings.TrimSpace(z.Text))
			for _, target := range lowerKW {
				if target == "confirm" {
					continue
				}
				if vision.FuzzySubstringMatch(txt, target, 1) {
					d.Logger.Info("🌀 Закрываем поп-ап", slog.String("popup", txt))
					if err := d.ADB.ClickRegion("ad_banner_close", d.AreaLookup); err != nil {
						d.Logger.Error("❌ Ошибка клика закрытия поп-апа", slog.Any("err", err))
						return err
					}
					time.Sleep(300 * time.Millisecond)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		// Если ни Confirm, ни поп-ап не сработали — ждём дальше
	}
}
