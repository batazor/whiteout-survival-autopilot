package imagefinder

import (
	"fmt"
	"image"
	"image/color"
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
func MatchIconInRegion(screenshotPath, iconPath string, region Region, threshold float32) (bool, float32, error) {
	// Load screenshot
	screenshot := gocv.IMRead(screenshotPath, gocv.IMReadColor)
	if screenshot.Empty() {
		return false, 0, ErrImageNotLoaded("screenshot")
	}
	defer screenshot.Close()

	// Load icon/template
	icon := gocv.IMRead(iconPath, gocv.IMReadColor)
	if icon.Empty() {
		return false, 0, ErrImageNotLoaded("icon")
	}
	defer icon.Close()

	// Ensure region is within bounds
	rect := image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	if rect.Max.X > screenshot.Cols() || rect.Max.Y > screenshot.Rows() {
		return false, 0, fmt.Errorf("region out of bounds: %+v", rect)
	}

	// Crop region from screenshot
	cropped := screenshot.Region(rect)
	defer cropped.Close()

	// Match template
	result := gocv.NewMat()
	defer result.Close()
	gocv.MatchTemplate(cropped, icon, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
	match := maxVal >= threshold

	// Debug output paths
	debugPath, heatmapPath := generateOutputPaths(screenshotPath)

	// Draw result rectangle if match found
	if match {
		topLeft := image.Pt(region.X+maxLoc.X, region.Y+maxLoc.Y)
		bottomRight := image.Pt(topLeft.X+icon.Cols(), topLeft.Y+icon.Rows())
		highlightColor := color.RGBA{G: 255, A: 255}
		gocv.Rectangle(&screenshot, image.Rect(topLeft.X, topLeft.Y, bottomRight.X, bottomRight.Y), highlightColor, 2)
	}

	// Save marked screenshot
	if ok := gocv.IMWrite(debugPath, screenshot); !ok {
		return false, maxVal, fmt.Errorf("failed to save debug image: %s", debugPath)
	}

	// Save heatmap image
	if err := saveHeatmap(result, heatmapPath); err != nil {
		return false, maxVal, err
	}

	return match, maxVal, nil
}

// saveHeatmap converts result matrix to heatmap image and saves it.
func saveHeatmap(result gocv.Mat, path string) error {
	// Normalize and convert to 8-bit
	gocv.Normalize(result, &result, 0, 255, gocv.NormMinMax)
	result8U := gocv.NewMat()
	defer result8U.Close()
	result.ConvertTo(&result8U, gocv.MatTypeCV8U)

	// Apply heatmap coloring
	heatmap := gocv.NewMat()
	defer heatmap.Close()
	gocv.ApplyColorMap(result8U, &heatmap, gocv.ColormapJet)

	if ok := gocv.IMWrite(path, heatmap); !ok {
		return fmt.Errorf("failed to save heatmap: %s", path)
	}
	return nil
}

// generateOutputPaths builds file paths for debug and heatmap outputs.
func generateOutputPaths(originalPath string) (debugPath string, heatmapPath string) {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(filepath.Base(originalPath), ext)
	dir := filepath.Dir(originalPath)

	debugPath = filepath.Join(dir, "debug_"+base+".png")
	heatmapPath = filepath.Join(dir, "debug_"+base+"_heatmap.png")
	return
}
