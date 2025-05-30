package analyzer

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
	"github.com/batazor/whiteout-survival-autopilot/internal/parser"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
)

type Analyzer struct {
	areas            *config.AreaLookup
	logger           *slog.Logger
	triggerEvaluator config.TriggerEvaluator
	usecaseLoader    config.UseCaseLoader
	ocrClient        *ocrclient.Client
}

func NewAnalyzer(areas *config.AreaLookup, logger *slog.Logger, ocrClient *ocrclient.Client) *Analyzer {
	return &Analyzer{
		areas:            areas,
		logger:           logger,
		triggerEvaluator: config.NewTriggerEvaluator(),
		usecaseLoader:    config.NewUseCaseLoader("./usecases"),
		ocrClient:        ocrClient,
	}
}

func (a *Analyzer) AnalyzeAndUpdateState(oldState *domain.Gamer, rules []domain.AnalyzeRule, queue *redis_queue.Queue) (*domain.Gamer, error) {
	newGamer := *oldState
	newChar := newGamer
	charPtr := &newChar

	// ========== 1️⃣ Делаем единый full-screen OCR ==========
	regions := make([]ocrclient.Region, 0)
	for _, rule := range rules {
		region, ok := a.areas.Get(rule.Name)
		if !ok {
			a.logger.Error("Region not found", slog.String("region", rule.Name))
			continue
		}

		regions = append(regions, ocrclient.Region{
			X0: region.Zone.Min.X,
			Y0: region.Zone.Min.Y,
			X1: region.Zone.Max.X,
			Y1: region.Zone.Max.Y,
		})
	}

	fullOCR, fullErr := a.ocrClient.FetchOCR("", regions) // debugName можно опустить
	if fullErr != nil {
		a.logger.Error("Full OCR failed", slog.Any("error", fullErr))
		return nil, fullErr
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, rule := range rules {
		wg.Add(1)

		go func(rule domain.AnalyzeRule) {
			defer wg.Done()

			threshold := rule.Threshold
			if threshold == 0 {
				threshold = 0.9
			}

			var value any

			switch rule.Action {
			case "exist":
				resp, err := a.ocrClient.FindImage(rule.Name, float64(threshold), rule.Name)
				if err != nil {
					a.logger.Error("FindImage failed",
						slog.String("image", rule.Name),
						slog.Any("error", err),
					)
					a.logger.Error("icon match failed", slog.String("region", rule.Name), slog.Any("error", err))
					return
				}
				value = resp.Found

			case "findIcon":
				// вызываем OCR-сервис — на Python-стороне к rule.Name автоматически добавят ".png"
				resp, err := a.ocrClient.FindImage(rule.Name, float64(threshold), rule.Name)
				if err != nil {
					a.logger.Error("FindImage failed",
						slog.String("icon", rule.Name),
						slog.Any("error", err),
					)
					return
				}

				// конвертируем полигоны в прямоугольники
				rects := resp.ToRects()
				matches := len(rects)
				a.logger.Info("📦 Icon search result",
					slog.String("icon", rule.Name),
					slog.Int("matches", matches),
				)
				value = resp.Found

				if rule.SaveAsRegion && resp.Found && matches > 0 {
					// берём лучший (первый) прямоугольник
					newBbox := rects[0]
					newRegion := config.Region{Zone: newBbox}
					a.areas.AddTemporaryRegion(rule.Name, newRegion)

					x, y := newBbox.Min.X, newBbox.Min.Y
					w, h := newBbox.Dx(), newBbox.Dy()
					a.logger.Info("💾 Saved new region from findIcon",
						slog.String("name", rule.Name),
						slog.Int("x", x),
						slog.Int("y", y),
						slog.Int("width", w),
						slog.Int("height", h),
					)
				}

			case "findText":
				if rule.Text == "" {
					a.logger.Warn("findText requires 'text' field", slog.String("rule", rule.Name))
					return
				}

				conf := rule.Threshold
				if conf == 0 {
					conf = 0.4
				}

				found := false
				var bbox domain.OCRResult
				for _, r := range fullOCR {
					if float64(r.Score)/100.0 < conf {
						continue
					}

					if strings.Contains(strings.ToLower(r.Text), strings.ToLower(rule.Text)) {
						found = true
						bbox = r
						break
					}
				}
				value = found

				if rule.SaveAsRegion && found {
					newRegion := config.Region{
						Zone: image.Rect(bbox.X, bbox.Y, bbox.X+bbox.Width, bbox.Y+bbox.Height),
					}
					a.areas.AddTemporaryRegion(rule.Name, newRegion)

					a.logger.Info("💾 Saved region from findText",
						slog.String("name", rule.Name),
						slog.Int("x", bbox.X),
						slog.Int("y", bbox.Y),
						slog.Int("w", bbox.Width),
						slog.Int("h", bbox.Height),
					)
				}

			case "color_check":
				zone, err := a.areas.GetRegionByName(rule.Name)
				if err != nil {
					a.logger.Error("GetRegionByName failed",
						slog.String("region", rule.Name),
						slog.Any("error", err),
					)
					return
				}

				ocrZoneResults := fullOCR.FilterByBBox(zone)

				if len(ocrZoneResults) == 0 {
					a.logger.Warn("No OCR results found in the specified region",
						slog.String("region", rule.Name),
						slog.String("expected_color_bg", rule.ExpectedColorBg),
						slog.String("expected_color_text", rule.ExpectedColorText),
					)
				}

				// проверяем, есть ли хотя бы одна зона с нужным цветом и достаточной уверенностью
				found := false
				for _, zr := range ocrZoneResults {
					if zr.Score < rule.Threshold {
						continue
					}
					if zr.AvgColor == rule.ExpectedColorText && rule.ExpectedColorText != "" {
						found = true
						break
					}
					if zr.BgColor == rule.ExpectedColorBg && rule.ExpectedColorBg != "" {
						found = true
						break
					}
				}
				value = found

			case "text":
				zone, err := a.areas.GetRegionByName(rule.Name)
				if err != nil {
					a.logger.Error("GetRegionByName failed",
						slog.String("region", rule.Name),
						slog.Any("error", err),
					)
					return
				}

				ocrZoneResults := fullOCR.FilterByBBox(zone)

				text := ""
				if len(ocrZoneResults) == 0 {
					a.logger.Warn("No OCR results found in the specified region",
						slog.String("region", rule.Name),
						slog.String("expected_text", rule.Text),
					)
				} else {
					text = ocrZoneResults[0].Text
				}

				a.logger.Info("text result", slog.String("region", rule.Name), slog.String("text", text))
				switch rule.Type {
				case "integer":
					value = parser.ParseNumber(text)
				case "string":
					value = text
				case "time_duration":
					value = parseTimeDuration(text)
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

			if value == nil {
				value = false
			}

			if err := setFieldByPath(reflect.ValueOf(charPtr).Elem(), strings.Split(rule.Name, "."), value); err != nil {
				panic(fmt.Sprintf("❌ failed to set field [%s]: %v", rule.Name, err))
			}
		}(rule)
	}

	wg.Wait()
	newGamer = *charPtr

	// Проверка pushUsecase'ов после установки значений
	if queue == nil {
		a.logger.Warn("❌ Queue is nil, skipping pushUsecase evaluation")
		return &newGamer, nil
	}

	for _, rule := range rules {
		for _, push := range rule.PushUseCase {
			// Проверяем что триггер выполняется
			if push.Trigger != "" {
				ok, err := a.triggerEvaluator.EvaluateTrigger(push.Trigger, charPtr)
				if err != nil {
					a.logger.Error("❌ Trigger evaluation failed for pushUsecase",
						slog.String("trigger", push.Trigger),
						slog.Any("error", err),
					)
					continue
				}
				if !ok {
					a.logger.Info("📭 Trigger not satisfied for pushUsecase",
						slog.String("trigger", push.Trigger),
						slog.String("currentState", newGamer.ScreenState.CurrentState),
					)
					continue
				}
			}

			// Если триггер выполнен, добавляем usecase в очередь
			for _, uc := range push.List {
				ucOriginal := a.usecaseLoader.GetByName(uc.Name)

				if ucOriginal == nil {
					a.logger.Error("❌ Usecase not found", slog.String("usecase", uc.Name))
					continue
				}

				a.logger.Info("📥 Push usecase from analysis", slog.String("usecase", uc.Name))
				if err := queue.Push(context.Background(), ucOriginal); err != nil {
					a.logger.Error("❌ Failed to push usecase", slog.String("usecase", uc.Name), slog.Any("error", err))
				}
			}
		}
	}

	return &newGamer, nil
}

// setFieldByPath sets a nested field by string path using reflection.
// Если value == false и поле целевого типа int/uint/string, ставит zero-value.
func setFieldByPath(v reflect.Value, path []string, value any) error {
	for i, part := range path {
		if i == len(path)-1 {
			// последний сегмент – непосредственно поле
			field := v.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, part)
			})
			if !field.IsValid() || !field.CanSet() {
				return fmt.Errorf("cannot set field: %s", part)
			}

			val := reflect.ValueOf(value)
			// если value == false и поле int или string — ставим zero-value
			if val.Kind() == reflect.Bool && !val.Bool() {
				switch field.Type().Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					field.SetInt(0)
					return nil
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					field.SetUint(0)
					return nil
				case reflect.String:
					field.SetString("")
					return nil
				}
			}
			// обычная попытка конвертации
			if val.Type().ConvertibleTo(field.Type()) {
				field.Set(val.Convert(field.Type()))
				return nil
			}
			return fmt.Errorf("type mismatch for field %s: cannot convert %s to %s",
				part, val.Type(), field.Type())
		}

		// идём по вложенным структурам / указателям
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
