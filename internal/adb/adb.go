package adb

import (
	"bytes"
	"fmt"
	"image"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// DeviceController defines the interface for interacting with an Android device via ADB.
type DeviceController interface {
	ListDevices() ([]string, error)
	SetActiveDevice(serial string)
	GetActiveDevice() string

	Screenshot(path string) (image.Image, error)
	ClickRegion(name string, area *config.AreaLookup) error
	ClickOCRResult(result *domain.OCRResult) error
	Swipe(x1 int, y1 int, x2 int, y2 int, durationMs time.Duration) error
}

// The Controller implements the DeviceController interface using the adb CLI tool.
type Controller struct {
	deviceID string
	logger   *slog.Logger
}

// NewController creates a new instance of the Controller.
func NewController(logger *slog.Logger, name string) (*Controller, error) {
	return &Controller{
		logger:   logger,
		deviceID: name,
	}, nil
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
func (a *Controller) Screenshot(path string) (image.Image, error) {
	a.logger.Info("Capturing screenshot from device",
		slog.String("device", a.deviceID),
		slog.String("output", path),
	)

	cmd := exec.Command("adb", "-s", a.deviceID, "exec-out", "screencap", "-p")
	out, err := cmd.Output()
	if err != nil {
		a.logger.Error("Failed to execute screencap", slog.Any("error", err))
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	if path != "" {
		if err := os.WriteFile(path, out, 0644); err != nil {
			a.logger.Error("Failed to write screenshot to file", slog.String("path", path), slog.Any("error", err))
			return nil, fmt.Errorf("failed to write screenshot: %w", err)
		}
		a.logger.Info("Screenshot saved successfully", slog.String("path", path))
	}

	img, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	return img, nil
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
	return cmd.Run()
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
	return cmd.Run()
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

	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "swipe",
		strconv.Itoa(startX), strconv.Itoa(startY),
		strconv.Itoa(endX), strconv.Itoa(endY),
		strconv.Itoa(int(durationMs.Milliseconds())),
	)

	a.logger.Info("Swipe with jitter",
		slog.Int("startX", startX),
		slog.Int("startY", startY),
		slog.Int("endX", endX),
		slog.Int("endY", endY),
		strconv.Itoa(int(durationMs.Milliseconds())),
	)

	return cmd.Run()
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
