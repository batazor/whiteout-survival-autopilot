package imagefinder

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

type Region struct {
	X      int
	Y      int
	Width  int
	Height int
}

// MatchIconInRegion performs template matching in a specified region of the screenshot.
// Returns:
//   - match: true if confidence >= threshold
//   - confidence: max match score from template matching
//   - error: if loading or processing failed
func MatchIconInRegion(screenshotPath, iconPath string, region Region, threshold float32, logger *slog.Logger) (bool, float32, error) {
	logger.Info("🔍 Starting icon match",
		slog.String("screenshot", screenshotPath),
		slog.String("icon", iconPath),
		slog.Any("region", region),
		slog.Float64("threshold", float64(threshold)),
	)

	screenshot := gocv.IMRead(screenshotPath, gocv.IMReadColor)
	if screenshot.Empty() {
		return false, 0, ErrImageNotLoaded(screenshotPath)
	}
	defer screenshot.Close()

	icon := gocv.IMRead(iconPath, gocv.IMReadColor)
	if icon.Empty() {
		return false, 0, ErrImageNotLoaded(iconPath)
	}
	defer icon.Close()

	rect := image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	if rect.Max.X > screenshot.Cols() || rect.Max.Y > screenshot.Rows() {
		return false, 0, fmt.Errorf("region out of bounds: %+v", rect)
	}

	cropped := screenshot.Region(rect)
	defer cropped.Close()

	result := gocv.NewMat()
	defer result.Close()

	// Use empty Mat instead of creating a new one that's never closed
	gocv.MatchTemplate(cropped, icon, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
	logger.Info("📊 Icon match result", slog.Float64("confidence", float64(maxVal)))
	match := maxVal >= threshold

	if match {
		topLeft := image.Pt(region.X+maxLoc.X, region.Y+maxLoc.Y)
		bottomRight := image.Pt(topLeft.X+icon.Cols(), topLeft.Y+icon.Rows())
		highlightColor := color.RGBA{G: 255, A: 255}
		gocv.Rectangle(&screenshot, image.Rect(topLeft.X, topLeft.Y, bottomRight.X, bottomRight.Y), highlightColor, 2)
	}

	// Save debug images
	debugPath, resultMapPath := generateOutputPaths(screenshotPath)
	if err := os.MkdirAll(filepath.Dir(debugPath), 0755); err != nil {
		logger.Warn("Failed to create debug directory", slog.String("error", err.Error()))
	}

	if ok := gocv.IMWrite(debugPath, screenshot); !ok {
		logger.Warn("Failed to save debug image", slog.String("path", debugPath))
	}

	grayscale := gocv.NewMat()
	defer grayscale.Close()
	gocv.Normalize(result, &grayscale, 0, 255, gocv.NormMinMax)

	grayscale8U := gocv.NewMat()
	defer grayscale8U.Close()
	grayscale.ConvertTo(&grayscale8U, gocv.MatTypeCV8U)

	if ok := gocv.IMWrite(resultMapPath, grayscale8U); !ok {
		logger.Warn("Failed to save result map", slog.String("path", resultMapPath))
	}

	return match, maxVal, nil
}

// generateOutputPaths builds file paths for debug and heatmap outputs.
func generateOutputPaths(originalPath string) (debugPath string, resultMapPath string) {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(filepath.Base(originalPath), ext)
	dir := filepath.Dir(originalPath)

	debugPath = filepath.Join(dir, "debug_"+base+".png")
	resultMapPath = filepath.Join(dir, "debug_"+base+"_matchmap.png")
	return
}
