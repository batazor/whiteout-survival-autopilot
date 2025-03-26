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

	// Optional: Validate actions
	for screen, ruleList := range rules {
		for i, rule := range ruleList {
			if rule.Action != "text" &&
				rule.Action != "exist" &&
				rule.Action != "color_check" {
				return nil, fmt.Errorf("invalid action '%s' in rule[%d] for screen '%s'", rule.Action, i, screen)
			}
		}
	}

	return rules, nil
}
