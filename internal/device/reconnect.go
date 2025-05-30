package device

//
//import (
//	"fmt"
//	"log/slog"
//	"strings"
//	"time"
//
//	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
//	"github.com/batazor/whiteout-survival-autopilot/internal/config"
//	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
//	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
//)
//
//type ReconnectHandler struct {
//	adbController adb.DeviceController
//	area          *config.AreaLookup
//	OCRClient     *ocrclient.Client
//	logger        *slog.Logger
//	maxAttempts   int
//}
//
//func (h *ReconnectHandler) HandleReconnect(screenshotPath string) error {
//	const waitAfterReconnectClick = 20 * time.Second
//	const waitAfterRestart = 10 * time.Second
//	const maxTimeout = 20 * time.Second
//
//	for restartCount := 0; ; restartCount++ {
//		attempt := 0
//
//		for attempt < h.maxAttempts {
//			attempt++
//
//			found, err := h.checkReconnectWindow(screenshotPath)
//			if err != nil {
//				h.logger.Warn("❌ Ошибка при проверке окна reconnect", slog.Any("err", err))
//				return err
//			}
//			if !found {
//				h.logger.Info("✅ Кнопка reconnect не обнаружена, продолжаем работу")
//				return nil
//			}
//
//			h.logger.Warn("Reconnect окно обнаружено, пробуем переподключиться",
//				slog.Int("attempt", attempt),
//				slog.Int("restartCount", restartCount),
//			)
//
//			if err := h.adbController.ClickRegion("reconnect_button", h.area); err != nil {
//				return fmt.Errorf("ошибка клика по кнопке reconnect: %w", err)
//			}
//
//			h.logger.Info("⏳ Ожидаем завершение загрузки после клика (20 секунд)")
//			time.Sleep(waitAfterReconnectClick)
//		}
//
//		// --- Рестарт приложения ---
//		h.logger.Error("🚨 Переподключение не удалось, перезапускаем приложение")
//		if err := h.adbController.RestartApplication(); err != nil {
//			return fmt.Errorf("не удалось перезапустить приложение: %w", err)
//		}
//
//		h.logger.Info("⏳ Ждем загрузку приложения (10 секунд)")
//		time.Sleep(waitAfterRestart)
//
//		// --- После рестарта ждём исчезновения кнопки reconnect ---
//		h.logger.Info("⏳ Ожидаем исчезновение кнопки reconnect (до 20 секунд)")
//
//		expire := time.After(maxTimeout)
//		tick := time.NewTicker(2 * time.Second)
//		defer tick.Stop()
//
//		for {
//			select {
//			case <-expire:
//				h.logger.Warn("🔁 Кнопка reconnect все еще на экране после рестарта — продолжаем цикл")
//				break
//
//			case <-tick.C:
//				found, err := h.checkReconnectWindow(screenshotPath)
//				if err != nil {
//					return err
//				}
//				if !found {
//					h.logger.Info("✅ Кнопка reconnect исчезла — продолжаем работу")
//					return nil
//				}
//			}
//		}
//	}
//}
//
//func (h *ReconnectHandler) checkReconnectWindow(screenshotPath string) (bool, error) {
//	// выполняем OCR только в области кнопки reconnect
//	results, err := h.OCRClient.FetchOCRByAreaName("reconnect_button", "reconnect_check")
//	if err != nil {
//		h.logger.Error("❌ OCRClient FetchOCRByAreaName failed for reconnect", slog.Any("error", err))
//		return false, err
//	}
//
//	// если ничего не распознано — окно не появилось
//	if len(results) == 0 {
//		return false, nil
//	}
//
//	// берём первый результат и приводим к нижнему регистру
//	text := strings.ToLower(strings.TrimSpace(results[0].Text))
//	h.logger.Info("🔍 OCR result reconnect", slog.String("text", text))
//
//	// нестрогий матч по слову "reconnect"
//	target := "reconnect"
//	if strings.Contains(text, target) || vision.FuzzySubstringMatch(text, target, 1) {
//		h.logger.Info("✅ reconnect window detected")
//		return true, nil
//	}
//
//	return false, nil
//}
