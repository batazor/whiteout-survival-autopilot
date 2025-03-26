package analyzer

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/imagefinder"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

type Analyzer struct {
	areas  *config.AreaLookup
	logger *slog.Logger
}

func NewAnalyzer(areas *config.AreaLookup, logger *slog.Logger) *Analyzer {
	return &Analyzer{
		areas:  areas,
		logger: logger,
	}
}

func (a *Analyzer) AnalyzeAndUpdateState(imagePath string, oldState *domain.State, rules []config.AnalyzeRule) (*domain.State, error) {
	if len(oldState.Accounts) == 0 || len(oldState.Accounts[0].Characters) == 0 {
		return nil, fmt.Errorf("no characters in state")
	}

	newState := *oldState
	newChar := newState.Accounts[0].Characters[0]
	charPtr := &newChar

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, rule := range rules {
		rule := rule // capture range variable
		wg.Add(1)

		go func() {
			defer wg.Done()

			bbox, err := a.areas.GetRegionByName(rule.Name)
			if err != nil {
				a.logger.Warn("region not found for rule", slog.String("region", rule.Name))
				return
			}

			region := bbox.ToRectangle()
			threshold := rule.Threshold
			if threshold == 0 {
				threshold = 0.9
			}

			var value any

			switch rule.Action {
			case "exist":
				iconPath := filepath.Join("references", "icons", filepath.Base(rule.Name)+".png")
				found, _, err := imagefinder.MatchIconInRegion(imagePath, iconPath, region, float32(threshold), a.logger)
				if err != nil {
					a.logger.Error("icon match failed", slog.String("region", rule.Name), slog.Any("error", err))
					return
				}
				value = found

			case "color_check":
				found, err := imagefinder.IsColorDominant(imagePath, region, rule.ExpectedColor, float32(threshold), a.logger)
				if err != nil {
					a.logger.Error("color check failed", slog.String("region", rule.Name), slog.Any("error", err))
					return
				}
				value = found

			case "text":
				text, err := vision.ExtractTextFromRegion(imagePath, region, rule.Name)
				if err != nil {
					a.logger.Error("OCR failed", slog.String("region", rule.Name), slog.Any("error", err))
					return
				}
				a.logger.Info("text result", slog.String("region", rule.Name), slog.String("text", text))
				switch rule.Type {
				case "integer":
					value = parseNumber(text)
				case "string":
					value = text
				default:
					a.logger.Warn("unsupported type", slog.String("type", rule.Type))
					return
				}
			default:
				a.logger.Warn("unsupported action", slog.String("action", rule.Action))
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if err := setFieldByPath(reflect.ValueOf(charPtr).Elem(), strings.Split(rule.Name, "."), value); err != nil {
				a.logger.Error("failed to set field", slog.String("path", rule.Name), slog.Any("error", err))
			}
		}()
	}

	wg.Wait()
	newState.Accounts[0].Characters[0] = *charPtr

	return &newState, nil
}

// parseNumber converts a string like "1 234 567" → 1234567
func parseNumber(s string) int {
	clean := strings.ReplaceAll(s, " ", "")
	val, _ := strconv.Atoi(clean)
	return val
}

// setFieldByPath sets a nested field by string path using reflection
func setFieldByPath(v reflect.Value, path []string, value any) error {
	for i, part := range path {
		if i == len(path)-1 {
			field := v.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, part)
			})
			if !field.IsValid() || !field.CanSet() {
				return fmt.Errorf("cannot set field: %s", part)
			}
			val := reflect.ValueOf(value)
			if val.Type().ConvertibleTo(field.Type()) {
				field.Set(val.Convert(field.Type()))
			} else {
				return fmt.Errorf("type mismatch for field %s", part)
			}
			return nil
		}

		v = v.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, part)
		})
		if !v.IsValid() {
			return fmt.Errorf("invalid field: %s", part)
		}
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
	}

	return nil
}
