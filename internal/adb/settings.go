package adb

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"

	"github.com/batazor/whiteout-survival-autopilot/internal/metrics"
)

// SetHeadsUpNotifications enables or disables heads-up notifications.
func (a *Controller) SetHeadsUpNotifications(enabled bool) error {
	value := "1"
	if !enabled {
		value = "0"
	}

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "settings", "put", "global", "heads_up_notifications_enabled", value)
	err := cmd.Run()
	if err != nil {
		a.logger.Error("Failed to set heads-up notifications", slog.Any("error", err), slog.String("value", value))
		metrics.ADBErrorTotal.WithLabelValues(a.deviceID, "heads_up_notifications").Inc()

		return fmt.Errorf("failed to set heads-up notifications: %w", err)
	}

	a.logger.Info("Heads-up notifications setting applied", slog.Bool("enabled", enabled))
	return nil
}

// SetBrightness sets the screen brightness on the device to the given percentage (0-100).
func (a *Controller) SetBrightness(percent int) error {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	value := int(float64(percent) / 100.0 * 255.0)

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "settings", "put", "system", "screen_brightness", strconv.Itoa(value))
	err := cmd.Run()
	if err != nil {
		a.logger.Error("Failed to set brightness", slog.Any("error", err), slog.Int("value", value))
		metrics.ADBErrorTotal.WithLabelValues(a.deviceID, "brightness").Inc()

		return fmt.Errorf("failed to set brightness: %w", err)
	}

	a.logger.Info("Brightness set successfully", slog.Int("value", value))
	return nil
}
