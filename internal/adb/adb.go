package adb

import (
	"fmt"
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

// ADBController implements the DeviceController interface using the adb CLI tool.
type ADBController struct {
	deviceID string
}

// NewADBController creates a new instance of the ADBController.
func NewADBController() *ADBController {
	return &ADBController{}
}

// ListDevices returns all connected ADB devices.
func (a *ADBController) ListDevices() ([]string, error) {
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
func (a *ADBController) SetActiveDevice(serial string) {
	a.deviceID = serial
}

// GetActiveDevice returns the currently selected device ID.
func (a *ADBController) GetActiveDevice() string {
	return a.deviceID
}

// Screenshot captures a screenshot from the active device and writes it to the given file path.
func (a *ADBController) Screenshot(path string) error {
	cmd := exec.Command("adb", "-s", a.deviceID, "exec-out", "screencap", "-p")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}

// ClickRegion performs a tap action in the center of the named region.
func (a *ADBController) ClickRegion(name string, area *config.AreaLookup) error {
	bbox, err := area.GetRegionByName(name)
	if err != nil {
		return fmt.Errorf("region '%s' not found: %w", name, err)
	}
	x, y, _, _ := bbox.ToPixels()
	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "tap",
		strconv.Itoa(x), strconv.Itoa(y),
	)
	return cmd.Run()
}

// Swipe performs a swipe gesture from (x1, y1) to (x2, y2) in the given duration (ms).
func (a *ADBController) Swipe(x1, y1, x2, y2, durationMs int) error {
	cmd := exec.Command("adb", "-s", a.deviceID, "shell", "input", "swipe",
		strconv.Itoa(x1), strconv.Itoa(y1),
		strconv.Itoa(x2), strconv.Itoa(y2),
		strconv.Itoa(durationMs),
	)
	return cmd.Run()
}
