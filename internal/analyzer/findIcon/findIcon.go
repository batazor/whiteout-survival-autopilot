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

func FindIcons(screenshotPath, iconPath string, threshold float32, logger *slog.Logger) (domain.BBoxes, error) {
	// Загружаем изображение один раз, чтобы получить размер
	img := gocv.IMRead(screenshotPath, gocv.IMReadGrayScale)
	if img.Empty() {
		return nil, fmt.Errorf("failed to load screenshot: %s", screenshotPath)
	}
	defer img.Close()

	fullRegion := image.Rect(0, 0, img.Cols(), img.Rows())

	return FindIconsInRegion(screenshotPath, iconPath, fullRegion, threshold, logger)
}

// FindIconsInRegion ищет иконки только в заданной области searchRegion (в координатах всего изображения).
func FindIconsInRegion(
	screenshotPath, iconPath string,
	searchRegion image.Rectangle,
	threshold float32,
	logger *slog.Logger,
) (domain.BBoxes, error) {
	// Загружаем скриншот полностью
	fullScreenshot := gocv.IMRead(screenshotPath, gocv.IMReadGrayScale)
	if fullScreenshot.Empty() {
		return nil, fmt.Errorf("failed to load screenshot: %s", screenshotPath)
	}
	defer fullScreenshot.Close()

	colorScreenshot := gocv.IMRead(screenshotPath, gocv.IMReadColor)
	if colorScreenshot.Empty() {
		return nil, fmt.Errorf("failed to load color screenshot: %s", screenshotPath)
	}
	defer colorScreenshot.Close()

	// Обрезаем регион
	screenshot := fullScreenshot.Region(searchRegion)
	defer screenshot.Close()

	// Загружаем иконку
	icon := gocv.IMRead(iconPath, gocv.IMReadGrayScale)
	if icon.Empty() {
		return nil, fmt.Errorf("failed to load icon: %s", iconPath)
	}
	defer icon.Close()

	result := gocv.NewMat()
	defer result.Close()

	gocv.MatchTemplate(screenshot, icon, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	var boxes []domain.BBox

	screenW := fullScreenshot.Cols()
	screenH := fullScreenshot.Rows()

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

		// Смещение от полной картинки
		absX := searchRegion.Min.X + maxLoc.X
		absY := searchRegion.Min.Y + maxLoc.Y

		bbox := domain.BBox{
			X:              float64(absX) / float64(screenW) * 100,
			Y:              float64(absY) / float64(screenH) * 100,
			Width:          float64(iconW) / float64(screenW) * 100,
			Height:         float64(iconH) / float64(screenH) * 100,
			Rotation:       0,
			OriginalWidth:  screenW,
			OriginalHeight: screenH,
			Confidence:     maxVal,
		}
		boxes = append(boxes, bbox)

		// Отрисовка на цветном скриншоте
		matchRect := image.Rect(absX, absY, absX+iconW, absY+iconH)
		gocv.Rectangle(&colorScreenshot, matchRect, color.RGBA{0, 255, 0, 255}, 2)

		// Затираем найденную область
		resultRect := image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+iconW, maxLoc.Y+iconH)
		gocv.Rectangle(&result, resultRect, color.RGBA{0, 0, 0, 0}, -1)
	}

	logger.Info("📦 Total matches found", slog.Int("count", len(boxes)))
	return boxes, nil
}

func generateDebugPath(original string) string {
	ext := filepath.Ext(original)
	name := strings.TrimSuffix(filepath.Base(original), ext)
	return filepath.Join("out", name+"_debug.png")
}
