package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

// LoadDeviceConfig читает YAML-файл конфигурации устройств и десериализует его в структуру domain.Config.
func LoadDeviceConfig(devicesFile string, repo repository.StateRepository) (*domain.Config, error) {
	ctx := context.Background()

	// 📄 Загружаем devices.yaml
	devicesData, err := os.ReadFile(filepath.Clean(devicesFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read devices.yaml: %w", err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(devicesData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices.yaml: %w", err)
	}

	// 🧠 Загружаем state из репозитория
	state, err := repo.LoadState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load state.yaml from repo: %w", err)
	}

	// 🔍 Индексируем state по gamer.ID
	stateMap := make(map[int]domain.Gamer)
	for _, g := range state.Gamers {
		stateMap[g.ID] = g
	}

	// 🔁 Мержим по ID и сортируем для стабильного порядка
	for dIdx := range cfg.Devices {
		for pIdx := range cfg.Devices[dIdx].Profiles {
			// 🔄 Мержим состояние для каждого игрока
			for gIdx, gamer := range cfg.Devices[dIdx].Profiles[pIdx].Gamer {
				if full, ok := stateMap[gamer.ID]; ok {
					cfg.Devices[dIdx].Profiles[pIdx].Gamer[gIdx] = full
				}
			}

			// 🔡 Сортируем игроков по Nickname
			sort.Sort(domain.Gamers(cfg.Devices[dIdx].Profiles[pIdx].Gamer))
		}

		// 📧 Сортируем профили по Email
		sort.Sort(domain.Profiles(cfg.Devices[dIdx].Profiles))
	}

	return &cfg, nil
}
