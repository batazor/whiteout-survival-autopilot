package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// LoadDeviceConfig читает YAML-файл конфигурации устройств и десериализует его в структуру domain.Config.
func LoadDeviceConfig(file string) (*domain.Config, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("failed to read device config file: %w", err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device config yaml: %w", err)
	}

	return &cfg, nil
}
