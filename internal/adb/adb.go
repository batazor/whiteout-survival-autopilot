package adb

import (
	"fmt"
	"image"
	"log/slog"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/metrics"
)

// DeviceController defines the interface for interacting with an Android device via ADB.
type DeviceController interface {
	ListDevices() ([]string, error)
	SetActiveDevice(serial string)
	GetActiveDevice() string
	RestartApplication() error

	Click(region image.Rectangle) error
	ClickRegion(name string, area *config.AreaLookup) error
	ClickOCRResult(result *domain.OCRResult) error

	Swipe(x1 int, y1 int, x2 int, y2 int, durationMs time.Duration) error
	SwipeDirection(direction string, delta int, durationMs time.Duration) error
}

// The Controller implements the DeviceController interface using the adb CLI tool.
type Controller struct {
	deviceID string
	logger   *slog.Logger
}

// NewController creates a new instance of the Controller.
func NewController(logger *slog.Logger, name string) (*Controller, error) {
	c := &Controller{
		logger:   logger,
		deviceID: name,
	}

	// Проверим доступность устройства
	if err := verifyDeviceAvailable(name); err != nil {
		panic(fmt.Sprintf("❌ %v", err))
	}

	// Set brightness to 70% on startup
	if err := c.SetBrightness(70); err != nil {
		logger.Warn("Failed to set initial brightness", slog.Any("error", err))
	}

	if err := c.SetHeadsUpNotifications(false); err != nil {
		logger.Warn("Failed to set initial heads-up notifications", slog.Any("error", err))
	}

	return c, nil
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

// ClickRegion performs a tap action in the center of the named region with slight random offset,
// clamping the result to stay inside the bounding box.
func (a *Controller) ClickRegion(name string, area *config.AreaLookup) error {
	bbox, err := area.GetRegionByName(name)
	if err != nil {
		return fmt.Errorf("region '%s' not found: %w", name, err)
	}

	x, y, w, h := bbox.ToPixels()

	centerX := x + w/2
	centerY := y + h/2

	offsetX := int(float64(w) * 0.05)
	offsetY := int(float64(h) * 0.05)

	randX := clamp(centerX+randInt(-offsetX, offsetX), x, x+w-1)
	randY := clamp(centerY+randInt(-offsetY, offsetY), y, y+h-1)

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "tap",
		strconv.Itoa(randX), strconv.Itoa(randY),
	)
	err = cmd.Run()
	if err != nil {
		a.logger.Error("Failed to execute tap command", slog.Any("error", err))
		metrics.ADBErrorTotal.WithLabelValues(a.deviceID, "click").Inc()

		return fmt.Errorf("failed to perform tap: %w", err)
	}

	return nil
}

// Click performs a tap action in the center of the given region with slight random offset,
// clamping the result to stay inside the bounding box.
func (a *Controller) Click(region image.Rectangle) error {
	x := region.Min.X
	y := region.Min.Y
	w := region.Dx()
	h := region.Dy()

	centerX := x + w/2
	centerY := y + h/2

	offsetX := int(float64(w) * 0.05)
	offsetY := int(float64(h) * 0.05)

	randX := clamp(centerX+randInt(-offsetX, offsetX), x, x+w-1)
	randY := clamp(centerY+randInt(-offsetY, offsetY), y, y+h-1)

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "tap",
		strconv.Itoa(randX), strconv.Itoa(randY),
	)
	err := cmd.Run()
	if err != nil {
		a.logger.Error("Failed to execute tap command", slog.Any("error", err))
		metrics.ADBErrorTotal.WithLabelValues(a.deviceID, "click").Inc()

		return fmt.Errorf("failed to perform tap: %w", err)
	}

	return nil
}

// ClickOCRResult performs a tap action in the center of the OCR result bounding box with slight random offset,
// clamping the result to stay inside the bounding box.
func (a *Controller) ClickOCRResult(result *domain.OCRResult) error {
	x, y, w, h := result.X, result.Y, result.Width, result.Height

	centerX := x + w/2
	centerY := y + h/2

	offsetX := int(float64(w) * 0.05)
	offsetY := int(float64(h) * 0.05)

	randX := clamp(centerX+randInt(-offsetX, offsetX), x, x+w-1)
	randY := clamp(centerY+randInt(-offsetY, offsetY), y, y+h-1)

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "tap",
		strconv.Itoa(randX), strconv.Itoa(randY),
	)
	err := cmd.Run()
	if err != nil {
		a.logger.Error("Failed to execute tap command", slog.Any("error", err))
		metrics.ADBErrorTotal.WithLabelValues(a.deviceID, "click").Inc()

		return fmt.Errorf("failed to perform tap: %w", err)
	}

	return nil
}

// randInt returns a random int in [min, max]
func randInt(min, max int) int {
	if min == max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

// Swipe performs a swipe gesture from (x1, y1) to (x2, y2) in the given duration (ms),
// adding slight randomness to simulate natural finger movement.
func (a *Controller) Swipe(x1 int, y1 int, x2 int, y2 int, durationMs time.Duration) error {
	// Добавим "дрожание" ±2 пикселя
	jitter := func(v int) int {
		return v + randInt(-2, 2)
	}

	startX := jitter(x1)
	startY := jitter(y1)
	endX := jitter(x2)
	endY := jitter(y2)

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "touchscreen", "swipe",
		strconv.Itoa(startX), strconv.Itoa(startY),
		strconv.Itoa(endX), strconv.Itoa(endY),
		strconv.Itoa(int(durationMs.Milliseconds())),
	)

	a.logger.Info("Swipe with jitter",
		slog.Int("startX", startX),
		slog.Int("startY", startY),
		slog.Int("endX", endX),
		slog.Int("endY", endY),
		slog.Duration("duration", durationMs),
	)

	err := cmd.Run()
	if err != nil {
		a.logger.Error("Failed to execute swipe command", slog.Any("error", err))
		metrics.ADBErrorTotal.WithLabelValues(a.deviceID, "swipe").Inc()

		return fmt.Errorf("failed to perform swipe: %w", err)
	}

	return nil
}

// LongTapRegion performs a long press in the center of the named region with jitter using the Swipe method.
func (a *Controller) LongTapRegion(name string, area *config.AreaLookup, durationMs time.Duration) error {
	bbox, err := area.GetRegionByName(name)
	if err != nil {
		return fmt.Errorf("region '%s' not found: %w", name, err)
	}

	x, y, w, h := bbox.ToPixels()
	centerX := x + w/2
	centerY := y + h/2

	a.logger.Info("Performing longtap via Swipe()",
		slog.String("region", name),
		slog.Int("x", centerX),
		slog.Int("y", centerY),
		slog.Duration("duration", durationMs),
	)

	// Просто используем Swipe с одинаковыми координатами и встроенным jitter
	return a.Swipe(centerX, centerY, centerX, centerY, durationMs)
}

// GetScreenResolution вызывает команду ADB shell "wm size",
// парсит результат и возвращает реальное разрешение экрана (width, height).
func (a *Controller) GetScreenResolution() (int, int, error) {
	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "wm", "size")
	out, err := cmd.Output()
	if err != nil {
		a.logger.Error("Failed to get screen resolution", slog.Any("error", err))
		return 0, 0, fmt.Errorf("failed to get screen resolution: %w", err)
	}

	// Ожидаемый формат вывода:
	// Physical size: 1080x2400
	// или
	// Override size: 1080x1920
	// Нужно найти подстроку вида "<num>x<num>"
	str := string(out)
	a.logger.Info("Raw wm size output", slog.String("output", str))

	// Ищем что-то вроде "1080x2400"
	var w, h int
	var matched bool

	lines := strings.Split(str, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Physical size:") ||
			strings.Contains(line, "Override size:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				res := parts[len(parts)-1] // "1080x2400"
				xy := strings.Split(res, "x")
				if len(xy) == 2 {
					ws, hs := xy[0], xy[1]
					wi, err1 := strconv.Atoi(ws)
					hi, err2 := strconv.Atoi(hs)
					if err1 == nil && err2 == nil {
						w, h = wi, hi
						matched = true
						break
					}
				}
			}
		}
	}

	if !matched {
		return 0, 0, fmt.Errorf("cannot parse screen resolution from wm size: %s", str)
	}

	a.logger.Info("Screen resolution found",
		slog.Int("width", w),
		slog.Int("height", h),
	)
	return w, h, nil
}

// clamp ограничивает значение в пределах [min, max]
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// verifyDeviceAvailable checks if the given ADB device is connected and ready.
func verifyDeviceAvailable(deviceID string) error {
	output, err := exec.Command("adb", "devices").Output()
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == deviceID && fields[1] == "device" {
			return nil
		}
	}

	return fmt.Errorf("ADB device '%s' not found or not in 'device' state", deviceID)
}

// SwipeDirection performs a swipe in the given direction (left, right, up, down) by delta pixels from the center of the screen.
func (a *Controller) SwipeDirection(direction string, delta int, durationMs time.Duration) error {
	width, height, err := a.GetScreenResolution()
	if err != nil {
		return fmt.Errorf("failed to get screen resolution: %w", err)
	}
	centerX, centerY := width/2, height/2

	var x1, y1, x2, y2 int
	switch strings.ToLower(direction) {
	case "left":
		x1, y1 = centerX, centerY
		x2, y2 = centerX-delta, centerY
	case "right":
		x1, y1 = centerX, centerY
		x2, y2 = centerX+delta, centerY
	case "up":
		x1, y1 = centerX, centerY
		x2, y2 = centerX, centerY-delta
	case "down":
		x1, y1 = centerX, centerY
		x2, y2 = centerX, centerY+delta
	default:
		return fmt.Errorf("unknown swipe direction: %s", direction)
	}

	a.logger.Info("SwipeDirection",
		slog.String("direction", direction),
		slog.Int("delta", delta),
		slog.Int("from_x", x1), slog.Int("from_y", y1),
		slog.Int("to_x", x2), slog.Int("to_y", y2),
		slog.Duration("duration", durationMs),
	)

	return a.Swipe(x1, y1, x2, y2, durationMs)
}
