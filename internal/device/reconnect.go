package device

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

// StartReconnectChecker запускает горутину проверки reconnect-окна в фоне
func (d *Device) StartReconnectChecker(ctx context.Context) {
	screenshotDir := filepath.Join("screenshots", d.Name)
	if err := os.MkdirAll(screenshotDir, os.ModePerm); err != nil {
		d.Logger.Error("❌ Не удалось создать директорию скриншотов", slog.Any("error", err))
		return
	}

	screenshotPath := filepath.Join(screenshotDir, "current.png")
	reconnectHandler := NewReconnectHandler(d.ADB, d.AreaLookup, d.Logger)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.Logger.Info("🛑 Остановка фоновой проверки reconnect")
			return
		case <-ticker.C:
			if err := reconnectHandler.HandleReconnect(screenshotPath); err != nil {
				d.Logger.Error("❌ Ошибка обработки reconnect", slog.Any("error", err))
			}
		}
	}
}

type ReconnectHandler struct {
	adbController adb.DeviceController
	area          *config.AreaLookup
	logger        *slog.Logger
	maxAttempts   int
}

func NewReconnectHandler(adb adb.DeviceController, area *config.AreaLookup, logger *slog.Logger) *ReconnectHandler {
	return &ReconnectHandler{
		adbController: adb,
		area:          area,
		logger:        logger,
		maxAttempts:   5,
	}
}

func (h *ReconnectHandler) HandleReconnect(screenshotPath string) error {
	const waitAfterReconnectClick = 20 * time.Second
	const waitAfterRestart = 10 * time.Second
	const maxTimeout = 20 * time.Second

	for restartCount := 0; ; restartCount++ {
		attempt := 0

		for attempt < h.maxAttempts {
			attempt++

			if _, err := h.adbController.Screenshot(screenshotPath); err != nil {
				h.logger.Warn("❌ Ошибка при снятии скриншота", slog.Any("err", err))
				return err
			}

			found, err := h.checkReconnectWindow(screenshotPath)
			if err != nil {
				h.logger.Warn("❌ Ошибка при проверке окна reconnect", slog.Any("err", err))
				return err
			}
			if !found {
				h.logger.Info("✅ Кнопка reconnect не обнаружена, продолжаем работу")
				return nil
			}

			h.logger.Warn("Reconnect окно обнаружено, пробуем переподключиться",
				slog.Int("attempt", attempt),
				slog.Int("restartCount", restartCount),
			)

			if err := h.adbController.ClickRegion("reconnect_button", h.area); err != nil {
				return fmt.Errorf("ошибка клика по кнопке reconnect: %w", err)
			}

			h.logger.Info("⏳ Ожидаем завершение загрузки после клика (20 секунд)")
			time.Sleep(waitAfterReconnectClick)
		}

		// --- Рестарт приложения ---
		h.logger.Error("🚨 Переподключение не удалось, перезапускаем приложение")
		if err := h.adbController.RestartApplication(); err != nil {
			return fmt.Errorf("не удалось перезапустить приложение: %w", err)
		}

		h.logger.Info("⏳ Ждем загрузку приложения (10 секунд)")
		time.Sleep(waitAfterRestart)

		// --- После рестарта ждём исчезновения кнопки reconnect ---
		h.logger.Info("⏳ Ожидаем исчезновение кнопки reconnect (до 20 секунд)")

		expire := time.After(maxTimeout)
		tick := time.NewTicker(2 * time.Second)
		defer tick.Stop()

		for {
			select {
			case <-expire:
				h.logger.Warn("🔁 Кнопка reconnect все еще на экране после рестарта — продолжаем цикл")
				break

			case <-tick.C:
				if _, err := h.adbController.Screenshot(screenshotPath); err != nil {
					return err
				}
				found, err := h.checkReconnectWindow(screenshotPath)
				if err != nil {
					return err
				}
				if !found {
					h.logger.Info("✅ Кнопка reconnect исчезла — продолжаем работу")
					return nil
				}
			}
		}
	}
}

func (h *ReconnectHandler) checkReconnectWindow(screenshotPath string) (bool, error) {
	region, err := h.area.GetRegionByName("reconnect_button")
	if err != nil {
		return false, fmt.Errorf("region reconnect_button not found: %w", err)
	}

	text, err := vision.ExtractTextFromRegion(
		screenshotPath,
		region.ToRectangle(),
		"reconnect_check",
		true, // использование CLAHE для улучшения OCR
	)
	if err != nil {
		h.logger.Error("❌ Ошибка OCR при проверке reconnect", slog.Any("error", err))
		return false, err
	}

	text = strings.ToLower(strings.TrimSpace(text))
	h.logger.Info("🔍 OCR результат reconnect", slog.String("text", text))

	// Нестрогое сравнение
	target := "reconnect"
	if strings.Contains(text, target) || vision.FuzzySubstringMatch(text, target, 1) {
		h.logger.Info("✅ Обнаружено окно reconnect")
		return true, nil
	}

	return false, nil
}

// CheckReconnectOnce выполняет проверку reconnect до исчезновения окна или рестарта
func (d *Device) CheckReconnectOnce(ctx context.Context) {
	d.Logger.Info("🔎 Проверка reconnect при старте устройства")

	screenshotDir := filepath.Join("screenshots", d.Name)
	if err := os.MkdirAll(screenshotDir, os.ModePerm); err != nil {
		d.Logger.Error("❌ Не удалось создать директорию скриншотов", slog.Any("error", err))
		return
	}

	screenshotPath := filepath.Join(screenshotDir, "current.png")
	reconnectHandler := NewReconnectHandler(d.ADB, d.AreaLookup, d.Logger)

	if err := reconnectHandler.HandleReconnect(screenshotPath); err != nil {
		d.Logger.Error("❌ Reconnect завершился с ошибкой", slog.Any("error", err))
	} else {
		d.Logger.Info("✅ Reconnect успешно обработан")
	}
}
