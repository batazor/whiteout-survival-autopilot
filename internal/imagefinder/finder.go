package imagefinder

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
)

// MatchIconInRegion performs template matching in a specified region of the screenshot.
// It automatically resizes the icon to fit the region while preserving aspect ratio,
// converts both to grayscale, applies Gaussian blur, and performs matching.
func MatchIconInRegion(
	screenshotPath string,
	iconPath string,
	region image.Rectangle,
	threshold float32,
	logger *logger.TracedLogger,
) (bool, float32, error) {
	ctx := context.Background()

	logger.Info(ctx, "🔍 Starting icon match",
		slog.String("screenshot", screenshotPath),
		slog.String("icon", iconPath),
		slog.Any("region", region),
		slog.Float64("threshold", float64(threshold)),
	)

	// Load screenshot
	screenshot := gocv.IMRead(screenshotPath, gocv.IMReadColor)
	if screenshot.Empty() {
		return false, 0, ErrImageNotLoaded("screenshot")
	}
	defer screenshot.Close()

	// Load icon/template
	icon := gocv.IMRead(iconPath, gocv.IMReadColor)
	if icon.Empty() {
		return false, 0, fmt.Errorf("failed to load icon from path: %s", iconPath)
	}
	defer icon.Close()

	// Ensure region is within bounds
	if region.Max.X > screenshot.Cols() || region.Max.Y > screenshot.Rows() {
		return false, 0, fmt.Errorf("region out of bounds: %+v", region)
	}

	// Crop region from screenshot
	cropped := screenshot.Region(region)
	defer cropped.Close()

	// Resize icon to fit cropped region while maintaining aspect ratio
	cropW := cropped.Cols()
	cropH := cropped.Rows()
	iconW := icon.Cols()
	iconH := icon.Rows()

	scaleX := float64(cropW) / float64(iconW)
	scaleY := float64(cropH) / float64(iconH)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newW := int(float64(iconW) * scale)
	newH := int(float64(iconH) * scale)

	if newW < 1 || newH < 1 {
		return false, 0, fmt.Errorf("scaled icon size too small")
	}

	resizedIcon := gocv.NewMat()
	defer resizedIcon.Close()
	gocv.Resize(icon, &resizedIcon, image.Pt(newW, newH), 0, 0, gocv.InterpolationArea)

	// Convert both to grayscale
	grayCrop := gocv.NewMat()
	grayIcon := gocv.NewMat()
	defer grayCrop.Close()
	defer grayIcon.Close()
	gocv.CvtColor(cropped, &grayCrop, gocv.ColorBGRToGray)
	gocv.CvtColor(resizedIcon, &grayIcon, gocv.ColorBGRToGray)

	// Apply Gaussian Blur
	gocv.GaussianBlur(grayCrop, &grayCrop, image.Pt(3, 3), 0, 0, gocv.BorderDefault)
	gocv.GaussianBlur(grayIcon, &grayIcon, image.Pt(3, 3), 0, 0, gocv.BorderDefault)

	// Template Matching
	result := gocv.NewMat()
	defer result.Close()
	gocv.MatchTemplate(grayCrop, grayIcon, &result, gocv.TmCcoeffNormed, gocv.NewMat())
	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)

	logger.Info(ctx, "📊 Icon match result", slog.Float64("confidence", float64(maxVal)))
	match := maxVal >= threshold

	// Highlight match in original screenshot
	if match {
		topLeft := image.Pt(region.Min.X+maxLoc.X, region.Min.Y+maxLoc.Y)
		bottomRight := image.Pt(topLeft.X+resizedIcon.Cols(), topLeft.Y+resizedIcon.Rows())
		highlightColor := color.RGBA{G: 255, A: 255}
		gocv.Rectangle(&screenshot, image.Rect(topLeft.X, topLeft.Y, bottomRight.X, bottomRight.Y), highlightColor, 2)
	}

	// Save annotated debug image and match map
	debugPath, resultMapPath := generateOutputPaths(screenshotPath)
	_ = os.MkdirAll(filepath.Dir(debugPath), 0755)
	_ = gocv.IMWrite(debugPath, screenshot)

	grayscale := gocv.NewMat()
	defer grayscale.Close()
	gocv.Normalize(result, &grayscale, 0, 255, gocv.NormMinMax)

	grayscale8U := gocv.NewMat()
	defer grayscale8U.Close()
	grayscale.ConvertTo(&grayscale8U, gocv.MatTypeCV8U)
	_ = gocv.IMWrite(resultMapPath, grayscale8U)

	return match, maxVal, nil
}

// generateOutputPaths builds file paths for debug and heatmap outputs.
func generateOutputPaths(originalPath string) (debugPath string, resultMapPath string) {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(filepath.Base(originalPath), ext)
	dir := "out"

	debugPath = filepath.Join(dir, "debug_"+base+".png")
	resultMapPath = filepath.Join(dir, "debug_"+base+"_matchmap.png")
	return
}
