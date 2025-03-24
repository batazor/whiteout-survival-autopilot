package analyzer

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/imagefinder"
	"github.com/batazor/whiteout-survival-autopilot/internal/utils"
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

		x, y, w, h := bbox.ToPixels()
		region := imagefinder.Region{X: x, Y: y, Width: w, Height: h}
		threshold := rule.Threshold
		if threshold == 0 {
			threshold = 0.9
		}

		switch rule.Action {
		case "exist":
			found, confidence, err := imagefinder.MatchIconInRegion(imagePath, rule.Name, region, float32(threshold))
			if err != nil {
				a.logger.Error("icon match failed", slog.String("region", rule.Name), slog.Any("error", err))
				continue
			}
			a.logger.Info("icon match result",
				slog.String("region", rule.Name),
				slog.Bool("found", found),
				slog.Float64("confidence", float64(confidence)),
			)

			if found {
				switch rule.Name {
				case "to_alliance":
					charPtr.Alliance.State.IsNeedSupport = true
				case "to_message":
					charPtr.Messages.State.IsNewMessage = true
				case "claim_button":
					charPtr.Messages.State.IsNewReports = true
				}
			}

		case "text":
			rect := bbox.ToRectangle()
			text, err := vision.ExtractTextFromRegion(imagePath, rect, rule.Name)
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

	// Log diff using lipgloss
	utils.PrintStyledDiff(oldState, &newState)

	return &newState, nil
}

// parseNumber converts a string like "1 234 567" → 1234567
func parseNumber(s string) int {
	clean := strings.ReplaceAll(s, " ", "")
	val, _ := strconv.Atoi(clean)
	return val
}
