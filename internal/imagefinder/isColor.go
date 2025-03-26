package imagefinder

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"strings"

	"gocv.io/x/gocv"
)

// IsColorDominant проверяет, доминирует ли ожидаемый цвет в указанном регионе изображения.
func IsColorDominant(imagePath string, region image.Rectangle, expected string, threshold float32, logger *slog.Logger) (bool, error) {
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return false, fmt.Errorf("failed to load image: %s", imagePath)
	}
	defer img.Close()

	crop := img.Region(region)
	defer crop.Close()

	mean := crop.Mean()
	blue := mean.Val1
	green := mean.Val2
	red := mean.Val3

	logger.Info("🧪 Checking color dominance",
		slog.String("image", imagePath),
		slog.Any("region", region),
		slog.String("expected", expected),
		slog.Float64("threshold_ratio", float64(threshold)),
		slog.Float64("threshold_absolute", float64(threshold*255)),
		slog.Float64("mean_blue", blue),
		slog.Float64("mean_green", green),
		slog.Float64("mean_red", red),
	)

	switch strings.ToLower(expected) {
	case "green":
		logger.Debug("Checking green dominance conditions",
			slog.Bool("green_gt_red+30", green > red+30),
			slog.Bool("green_gt_blue+30", green > blue+30),
			slog.Bool("green_gt_threshold", green > float64(threshold*255)),
		)
		if green > red+30 && green > blue+30 && green > float64(threshold*255) {
			logger.Info("✅ Green is dominant")
			return true, nil
		}
	case "red":
		logger.Debug("Checking red dominance conditions",
			slog.Bool("red_gt_green+30", red > green+30),
			slog.Bool("red_gt_blue+30", red > blue+30),
			slog.Bool("red_gt_threshold", red > float64(threshold*255)),
		)
		if red > green+30 && red > blue+30 && red > float64(threshold*255) {
			logger.Info("✅ Red is dominant")
			return true, nil
		}
	case "blue":
		logger.Debug("Checking blue dominance conditions",
			slog.Bool("blue_gt_red+30", blue > red+30),
			slog.Bool("blue_gt_green+30", blue > green+30),
			slog.Bool("blue_gt_threshold", blue > float64(threshold*255)),
		)
		if blue > red+30 && blue > green+30 && blue > float64(threshold*255) {
			logger.Info("✅ Blue is dominant")
			return true, nil
		}
	case "gray":
		diff1 := math.Abs(blue - green)
		diff2 := math.Abs(green - red)
		diff3 := math.Abs(blue - red)
		logger.Debug("Checking gray conditions",
			slog.Float64("diff_blue_green", diff1),
			slog.Float64("diff_green_red", diff2),
			slog.Float64("diff_blue_red", diff3),
			slog.Bool("blue_lt_200", blue < 200),
		)
		if diff1 < 15 && diff2 < 15 && diff3 < 15 && blue < 200 {
			logger.Info("✅ Gray is dominant")
			return true, nil
		}
	default:
		logger.Error("❌ Unsupported color", slog.String("color", expected))
		return false, fmt.Errorf("unsupported expected color: %s", expected)
	}

	logger.Info("❌ Expected color is not dominant")
	return false, nil
}
