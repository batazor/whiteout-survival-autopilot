package imagefinder

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"strings"

	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

func IsColorDominant(
	imagePath string,
	region image.Rectangle,
	expected string,
	ratioThreshold float32,
	logger *slog.Logger,
) (bool, error) {

	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return false, fmt.Errorf("failed to load image: %s", imagePath)
	}
	defer img.Close()

	ratio, err := PixelDominance(img, region, expected, ratioThreshold)
	if err != nil {
		return false, err
	}

	logger.Info("PixelDominance",
		slog.String("color", expected),
		slog.Float64("ratio", float64(ratio)),
		slog.Float64("threshold", float64(ratioThreshold)),
	)

	return ratio >= ratioThreshold, nil
}

// CheckRegionColor делает скриншот и проверяет доминирование цвета в заданной зоне.
func CheckRegionColor(
	ctx context.Context,
	adb adb.DeviceController,
	lookup *config.AreaLookup,
	regionName string,
	expectedColor string,
	threshold float32,
	logger *slog.Logger,
) (bool, error) {
	logger.Info("📸 Делаем скриншот для анализа цвета",
		slog.String("region", regionName),
		slog.String("expected_color", expectedColor),
		slog.Float64("threshold", float64(threshold)),
	)

	imagePath := fmt.Sprintf("screenshots/check_%s.png", regionName)

	_, err := adb.Screenshot(imagePath)
	if err != nil {
		logger.Error("❌ Не удалось сделать скриншот", slog.Any("err", err))
		return false, err
	}

	region, ok := lookup.Get(regionName)
	if !ok {
		return false, fmt.Errorf("region '%s' not found in area definitions", regionName)
	}

	// Добавляем информацию о регионе в логгер
	newLogger := logger.With(
		slog.String("region", regionName),
	)

	result, err := IsColorDominant(imagePath, region.Zone, expectedColor, threshold, newLogger)
	if err != nil {
		return false, err
	}

	logger.Info("🎨 Результат анализа цвета",
		slog.String("region", regionName),
		slog.String("expected_color", expectedColor),
		slog.Bool("is_dominant", result),
	)

	return result, nil
}

func PixelDominance(
	img gocv.Mat,
	region image.Rectangle,
	colorName string,
	ratioThreshold float32,
) (float32, error) {

	spec, ok := colorSpecs[strings.ToLower(colorName)]
	if !ok {
		return 0, fmt.Errorf("unsupported color '%s'", colorName)
	}

	crop := img.Region(region)
	defer crop.Close()

	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(crop, &hsv, gocv.ColorBGRToHSV)

	h, s, v := gocv.Split(hsv)[0], gocv.Split(hsv)[1], gocv.Split(hsv)[2]
	defer h.Close()
	defer s.Close()
	defer v.Close()

	rows, cols := hsv.Rows(), hsv.Cols()
	var total, match int

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			H := float32(h.GetUCharAt(y, x))         // 0‑179
			S := float32(s.GetUCharAt(y, x)) / 255.0 // 0‑1
			V := float32(v.GetUCharAt(y, x)) / 255.0 // 0‑1

			if S < spec.MinSat || S > spec.MaxSat ||
				V < spec.MinVal || V > spec.MaxVal {
				continue
			}
			total++

			for _, rng := range spec.HueRanges {
				if H >= rng[0] && H <= rng[1] {
					match++
					break
				}
			}
		}
	}

	if total == 0 {
		// Ни одного пикселя не прошло фильтр: считаем долю 0, ошибки нет.
		return 0, nil
	}

	return float32(match) / float32(total), nil
}
