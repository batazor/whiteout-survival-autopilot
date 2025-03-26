package adb

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

// DeviceController defines the interface for interacting with an Android device via ADB.
type DeviceController interface {
	ListDevices() ([]string, error)
	SetActiveDevice(serial string)
	GetActiveDevice() string

	Screenshot(path string) error
	ClickRegion(name string, area *config.AreaLookup) error
	Swipe(x1, y1, x2, y2, durationMs int) error
}

// The Controller implements the DeviceController interface using the adb CLI tool.
type Controller struct {
	deviceID string
	logger   *slog.Logger
}

// NewADBController creates a new instance of the Controller.
func NewADBController(logger *slog.Logger) *Controller {
	return &Controller{
		logger: logger,
	}
}

// ListDevices returns all connected ADB devices.
func (a *Controller) ListDevices() ([]string, error) {
	out, err := exec.Command("adb", "devices").Output()
	if err != nil {
		return nil, fmt.Errorf("adb not found or failed to list devices: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var result []string
	for _, line := range lines[1:] {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "device" {
			result = append(result, parts[0])
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no devices connected")
	}
	return result, nil
}

// SetActiveDevice sets the device ID to be used for all ADB commands.
func (a *Controller) SetActiveDevice(serial string) {
	a.deviceID = serial
}

// GetActiveDevice returns the currently selected device ID.
func (a *Controller) GetActiveDevice() string {
	return a.deviceID
}

// Screenshot captures a screenshot from the active device and writes it to the given file path.
func (a *Controller) Screenshot(path string) error {
	a.logger.Info("Capturing screenshot from device",
		slog.String("device", a.deviceID),
		slog.String("output", path),
	)

	cmd := exec.Command("adb", "-s", a.deviceID, "exec-out", "screencap", "-p")
	out, err := cmd.Output()
	if err != nil {
		a.logger.Error("Failed to execute screencap", slog.Any("error", err))
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		a.logger.Error("Failed to write screenshot to file", slog.String("path", path), slog.Any("error", err))
		return fmt.Errorf("failed to write screenshot: %w", err)
	}

	a.logger.Info("Screenshot saved successfully", slog.String("path", path))
	return nil
}

// ClickRegion performs a tap action in the center of the named region with slight random offset.
func (a *Controller) ClickRegion(name string, area *config.AreaLookup) error {
	bbox, err := area.GetRegionByName(name)
	if err != nil {
		return fmt.Errorf("region '%s' not found: %w", name, err)
	}

	x, y, w, h := bbox.ToPixels()

	// Центр области
	centerX := x + w/2
	centerY := y + h/2

	// Отклонение до 5% от ширины и высоты
	offsetX := int(float64(w) * 0.05)
	offsetY := int(float64(h) * 0.05)

	// Генерация случайного отклонения в диапазоне [-offsetX..offsetX]
	randX := centerX + randInt(-offsetX, offsetX)
	randY := centerY + randInt(-offsetY, offsetY)

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "tap",
		strconv.Itoa(randX), strconv.Itoa(randY),
	)
	return cmd.Run()
}

// randInt returns a random int in [min, max]
func randInt(min, max int) int {
	if min == max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

// Swipe performs a swipe gesture from (x1, y1) to (x2, y2) in the given duration (ms).
func (a *Controller) Swipe(x1, y1, x2, y2, durationMs int) error {
	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "swipe",
		strconv.Itoa(x1), strconv.Itoa(y1),
		strconv.Itoa(x2), strconv.Itoa(y2),
		strconv.Itoa(durationMs),
	)
	return cmd.Run()
}
