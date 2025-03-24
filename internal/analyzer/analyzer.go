package analyzer

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/imagefinder"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

type Analyzer struct {
	areas  *config.AreaLookup
	rules  config.ScreenAnalyzeRules
	logger *slog.Logger
}

func NewAnalyzer(areas *config.AreaLookup, rules config.ScreenAnalyzeRules, logger *slog.Logger) *Analyzer {
	return &Analyzer{
		areas:  areas,
		rules:  rules,
		logger: logger,
	}
}

func (a *Analyzer) AnalyzeAndUpdateState(imagePath string, oldState *domain.State, screen string) (*domain.State, error) {
	rules, ok := a.rules[screen]
	if !ok {
		a.logger.Warn("no analysis rules found for screen", slog.String("screen", screen))
		return oldState, nil
	}

	if len(oldState.Accounts) == 0 || len(oldState.Accounts[0].Characters) == 0 {
		return nil, fmt.Errorf("no characters in state")
	}

	newState := *oldState
	newChar := newState.Accounts[0].Characters[0]
	charPtr := &newChar

	for _, rule := range rules {
		bbox, err := a.areas.GetRegionByName(rule.Name)
		if err != nil {
			a.logger.Warn("region not found for rule", slog.String("region", rule.Name))
			continue
		}

		region := bbox.ToRectangle()
		threshold := rule.Threshold
		if threshold == 0 {
			threshold = 0.9
		}

		switch rule.Action {
		case "exist":
			iconPath := filepath.Join("references", "icons", rule.Name+".png")
			found, confidence, err := imagefinder.MatchIconInRegion(
				imagePath,
				iconPath,
				region,
				float32(threshold),
				a.logger,
			)
			if err != nil {
				a.logger.Error("icon match failed",
					slog.String("region", rule.Name),
					slog.Any("error", err),
					slog.String("image_path", imagePath),
				)
				continue
			}
			a.logger.Info("icon match result",
				slog.String("region", rule.Name),
				slog.Bool("found", found),
				slog.Float64("confidence", float64(confidence)),
			)

			switch rule.Name {
			case "allience_help":
				charPtr.Alliance.State.IsNeedSupport = found
			case "to_message":
				charPtr.Messages.State.IsNewMessage = found
			case "claim_button":
				charPtr.Messages.State.IsNewReports = found
			}

		case "text":
			text, err := vision.ExtractTextFromRegion(imagePath, region, rule.Name)
			if err != nil {
				a.logger.Error("OCR failed", slog.String("region", rule.Name), slog.Any("error", err))
				continue
			}

			a.logger.Info("text result",
				slog.String("region", rule.Name),
				slog.String("text", text),
			)

			switch rule.Name {
			case "power":
				val := parseNumber(text)
				charPtr.Power = val
			case "vipLevel":
				val := parseNumber(text)
				charPtr.VIPLevel = val
			}
		}
	}

	newState.Accounts[0].Characters[0] = *charPtr

	return &newState, nil
}

// parseNumber converts a string like "1 234 567" → 1234567
func parseNumber(s string) int {
	clean := strings.ReplaceAll(s, " ", "")
	val, _ := strconv.Atoi(clean)
	return val
}
