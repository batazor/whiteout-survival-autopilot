package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func LoadStateSnapshot(file string) (map[int]domain.Gamer, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("failed to read state.yaml: %w", err)
	}

	var snapshot domain.State
	if err := yaml.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state.yaml: %w", err)
	}

	stateMap := make(map[int]domain.Gamer)
	for _, account := range snapshot.Accounts {
		for _, g := range account.Characters {
			stateMap[g.ID] = g
		}
	}

	return stateMap, nil
}
