package teaapp

import (
	"fmt"
	"log/slog"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
)

func InitADBController(logger *slog.Logger) (adb.DeviceController, error) {
	controller := adb.NewController(logger)

	devices, err := controller.ListDevices()
	if err != nil {
		return nil, fmt.Errorf("ADB error: %w", err)
	}

	switch len(devices) {
	case 0:
		return nil, fmt.Errorf("no ADB devices connected")
	case 1:
		controller.SetActiveDevice(devices[0])
		logger.Info("Single ADB device selected", slog.String("device", devices[0]))
		return controller, nil
	default:
		logger.Warn("Multiple ADB devices found", slog.Int("count", len(devices)))
		return controller, nil // UI selection will be handled in Run()
	}
}
