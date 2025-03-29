package finder

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func FindIcons(screenshotPath, iconPath string, threshold float32, logger *slog.Logger) ([]domain.BBox, error) {
	screenshot := gocv.IMRead(screenshotPath, gocv.IMReadGrayScale)
	if screenshot.Empty() {
		return nil, fmt.Errorf("failed to load screenshot: %s", screenshotPath)
	}
	defer screenshot.Close()

	// Цветная версия для отрисовки
	colorScreenshot := gocv.IMRead(screenshotPath, gocv.IMReadColor)
	if colorScreenshot.Empty() {
		return nil, fmt.Errorf("failed to load color screenshot: %s", screenshotPath)
	}
	defer colorScreenshot.Close()

	icon := gocv.IMRead(iconPath, gocv.IMReadGrayScale)
	if icon.Empty() {
		return nil, fmt.Errorf("failed to load icon: %s", iconPath)
	}
	defer icon.Close()

	result := gocv.NewMat()
	defer result.Close()

	gocv.MatchTemplate(screenshot, icon, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	var boxes []domain.BBox

	sWidth := screenshot.Cols()
	sHeight := screenshot.Rows()

	iconW := icon.Cols()
	iconH := icon.Rows()

	for {
		_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)

		logger.Debug("🎯 MatchTemplate result",
			slog.Float64("confidence", float64(maxVal)),
			slog.Int("x", maxLoc.X),
			slog.Int("y", maxLoc.Y),
		)

		if maxVal < threshold {
			break
		}

		bbox := domain.BBox{
			X:              float64(maxLoc.X) / float64(sWidth) * 100,
			Y:              float64(maxLoc.Y) / float64(sHeight) * 100,
			Width:          float64(iconW) / float64(sWidth) * 100,
			Height:         float64(iconH) / float64(sHeight) * 100,
			Rotation:       0,
			OriginalWidth:  sWidth,
			OriginalHeight: sHeight,
		}
		boxes = append(boxes, bbox)

		// Отрисовка прямоугольника на цветной версии
		matchRect := image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+iconW, maxLoc.Y+iconH)
		gocv.Rectangle(&colorScreenshot, matchRect, color.RGBA{0, 255, 0, 255}, 2)

		// "Затираем" найденную область
		gocv.Rectangle(&result, matchRect, color.RGBA{0, 0, 0, 0}, -1)
	}

	// Сохраняем debug-изображение
	if len(boxes) > 0 {
		debugPath := generateDebugPath(screenshotPath)
		if err := gocv.IMWrite(debugPath, colorScreenshot); err == false {
			logger.Info("🖼️ Debug image saved", slog.String("path", debugPath))
		} else {
			logger.Warn("failed to save debug image", slog.String("path", debugPath), slog.Any("error", err))
		}
	}

	logger.Info("📦 Total matches found", slog.Int("count", len(boxes)))
	return boxes, nil
}

func generateDebugPath(original string) string {
	ext := filepath.Ext(original)
	name := strings.TrimSuffix(filepath.Base(original), ext)
	return filepath.Join("out", name+"_debug.png")
}
