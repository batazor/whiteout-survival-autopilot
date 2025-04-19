package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type ScreenAnalyzeRules map[string][]domain.AnalyzeRule

func LoadAnalyzeRules(path string) (ScreenAnalyzeRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read analyze config: %w", err)
	}

	var rules ScreenAnalyzeRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse analyze config: %w", err)
	}

	// Validate actions
	for screen, ruleList := range rules {
		for _, rule := range ruleList {
			if err := rule.Validate(); err != nil {
				return nil, fmt.Errorf("screen '%s': %w", screen, err)
			}
		}
	}

	return rules, nil
}
