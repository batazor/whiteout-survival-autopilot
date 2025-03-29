package analyzer

import (
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	finder "github.com/batazor/whiteout-survival-autopilot/internal/analyzer/findIcon"
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

func (a *Analyzer) AnalyzeAndUpdateState(imagePath string, oldState *domain.State, rules []domain.AnalyzeRule) (*domain.State, error) {
	if len(oldState.Accounts) == 0 || len(oldState.Accounts[0].Characters) == 0 {
		return nil, fmt.Errorf("no characters in state")
	}

	for _, rule := range rules {
		a.logger.Info("🧪 DSL rule",
			slog.String("name", rule.Name),
			slog.String("action", rule.Action),
			slog.String("expectedColor", rule.ExpectedColor),
		)
	}

	newState := *oldState
	newChar := newState.Accounts[0].Characters[0]
	charPtr := &newChar

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, rule := range rules {
		rule := rule // capture range variable

		// 🔍 Логируем содержимое правила
		a.logger.Info("🔎 AnalyzeRule loaded",
			slog.String("name", rule.Name),
			slog.String("action", rule.Action),
			slog.String("type", rule.Type),
			slog.Float64("threshold", rule.Threshold),
			slog.String("expectedColor", rule.ExpectedColor),
			slog.Bool("saveAsRegion", rule.SaveAsRegion),
		)

		wg.Add(1)

		go func() {
			defer wg.Done()

			var region image.Rectangle
			bbox, err := a.areas.GetRegionByName(rule.Name)
			if err != nil {
				if rule.SaveAsRegion {
					a.logger.Warn("region not found for rule (will try to detect and save)", slog.String("region", rule.Name))
					region = image.Rect(0, 0, 1080, 2400) // fallback: весь экран
				} else {
					panic(fmt.Sprintf("❌ Region not found for rule: %s", rule.Name))
				}
			} else {
				region = bbox.ToRectangle()
			}

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

			case "findIcon":
				iconPath := filepath.Join("references", "icons", filepath.Base(rule.Name)+".png")
				a.logger.Info("🔎 Starting findIcon", slog.String("rule", rule.Name), slog.String("iconPath", iconPath), slog.Float64("threshold", float64(threshold)))

				boxes, err := finder.FindIcons(imagePath, iconPath, float32(threshold), a.logger)
				if err != nil {
					a.logger.Error("❌ Icon search failed", slog.String("icon", rule.Name), slog.Any("error", err))
					return
				}

				a.logger.Info("📦 Icon search result", slog.String("icon", rule.Name), slog.Int("matches", len(boxes)))

				value = len(boxes) > 0

				if rule.SaveAsRegion && len(boxes) > 0 {
					bbox := boxes[0]
					x, y, w, h := bbox.ToPixels()
					newRegion := config.Region{Zone: image.Rect(x, y, x+w, y+h)}
					a.areas.AddTemporaryRegion(rule.Name, newRegion)

					a.logger.Info("💾 Saved new region from findIcon",
						slog.String("name", rule.Name),
						slog.Int("x", x),
						slog.Int("y", y),
						slog.Int("width", w),
						slog.Int("height", h),
					)
				}

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
				panic(fmt.Sprintf("❌ failed to set field [%s]: %v", rule.Name, err))
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
