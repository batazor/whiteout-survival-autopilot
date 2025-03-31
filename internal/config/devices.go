package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// LoadDeviceConfig читает YAML-файл конфигурации устройств и десериализует его в структуру domain.Config.
func LoadDeviceConfig(devicesFile string) (*domain.Config, error) {
	devicesData, err := os.ReadFile(filepath.Clean(devicesFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read devices.yaml: %w", err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(devicesData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices.yaml: %w", err)
	}

	// Загрузка state.yaml
	stateMap, err := LoadStateSnapshot("db/state.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load state.yaml: %w", err)
	}

	// Мержим по ID
	for dIdx := range cfg.Devices {
		for pIdx := range cfg.Devices[dIdx].Profiles {
			for gIdx, gamer := range cfg.Devices[dIdx].Profiles[pIdx].Gamer {
				if full, ok := stateMap[gamer.ID]; ok {
					cfg.Devices[dIdx].Profiles[pIdx].Gamer[gIdx] = full
				}
			}
		}
	}

	return &cfg, nil
}
