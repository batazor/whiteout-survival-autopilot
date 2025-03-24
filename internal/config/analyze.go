package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AnalyzeRule struct {
	Name      string  `yaml:"name"`      // Название региона (например: power, to_message)
	Action    string  `yaml:"action"`    // Действие: "text" или "exist"
	Type      string  `yaml:"type"`      // Тип значения: "integer", "string" и т.д. (для action: text)
	Threshold float64 `yaml:"threshold"` // Порог уверенности (например: 0.9), опционально
}

type ScreenAnalyzeRules map[string][]AnalyzeRule

func LoadAnalyzeRules(path string) (ScreenAnalyzeRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read analyze config: %w", err)
	}

	var rules ScreenAnalyzeRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse analyze config: %w", err)
	}

	return rules, nil
}
