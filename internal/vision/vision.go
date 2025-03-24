package vision

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

// ExtractTextFromRegion performs OCR on a specific screen region and highlights text boxes.
func ExtractTextFromRegion(imagePath string, zone image.Rectangle) (string, error) {
	// Read image
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return "", fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	// Crop the region
	cropped := img.Region(zone)
	defer cropped.Close()

	// Save for OCR
	tmpCropPath := filepath.Join(os.TempDir(), "ocr_crop.png")
	if !gocv.IMWrite(tmpCropPath, cropped) {
		return "", fmt.Errorf("failed to write cropped image")
	}

	// OCR with tesseract
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(tmpCropPath)
	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %w", err)
	}

	// Get bounding boxes
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return text, fmt.Errorf("failed to get bounding boxes: %w", err)
	}

	// Draw boxes on original image
	for _, box := range boxes {
		pt1 := image.Pt(zone.Min.X+box.Box.Min.X, zone.Min.Y+box.Box.Min.Y)
		pt2 := image.Pt(zone.Min.X+box.Box.Max.X, zone.Min.Y+box.Box.Max.Y)
		gocv.Rectangle(&img, image.Rect(pt1.X, pt1.Y, pt2.X, pt2.Y), color.RGBA{255, 0, 0, 255}, 2)
	}

	// Save annotated image
	outPath := filepath.Join("out", "ocr_annotated.png")
	if !gocv.IMWrite(outPath, img) {
		return text, fmt.Errorf("failed to save annotated image")
	}

	// Open image
	if err := openImage(outPath); err != nil {
		return text, fmt.Errorf("failed to open image: %w", err)
	}

	return text, nil
}

func openImage(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
