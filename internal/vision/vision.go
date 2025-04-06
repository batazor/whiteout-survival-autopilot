package vision

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

// ExtractTextFromRegion performs OCR on a specific screen region and highlights text boxes.
func ExtractTextFromRegion(imagePath string, zone image.Rectangle, outputName string, clane bool) (string, error) {
	// Load full screenshot
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return "", fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	// Crop to region
	cropped := img.Region(zone)
	defer cropped.Close()

	// Preprocess: grayscale -> threshold
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(cropped, &gray, gocv.ColorBGRToGray)

	// Optional: CLAHE
	if clane {
		clahe := gocv.NewCLAHE()
		defer clahe.Close()
		clahe.Apply(gray, &gray)
	}

	// Thresholding
	bin := gocv.NewMat()
	defer bin.Close()
	gocv.Threshold(gray, &bin, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	// Объединяет символы с разрывами
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(2, 2))
	gocv.MorphologyEx(bin, &bin, gocv.MorphClose, kernel)

	// Optional: Resize up if too small
	if bin.Cols() < 100 {
		scale := 2.0
		resized := gocv.NewMat()
		defer resized.Close()
		gocv.Resize(bin, &resized, image.Point{}, scale, scale, gocv.InterpolationLinear)
		bin = resized.Clone()
	}

	// Save preprocessed for debug
	preOut := filepath.Join("out", fmt.Sprintf("ocr_preprocessed_%s.png", outputName))
	_ = gocv.IMWrite(preOut, bin)

	// Save temp file for Tesseract
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("ocr_crop_%s.png", outputName))
	if !gocv.IMWrite(tmpPath, bin) {
		return "", fmt.Errorf("failed to write cropped image")
	}

	// OCR
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(tmpPath)
	client.SetLanguage("eng")
	client.SetWhitelist("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ[]")
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)
	client.SetVariable("user_defined_dpi", "814") // 407 ppi * 2

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %w", err)
	}

	// Annotate bounding boxes on original image
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err == nil {
		for _, box := range boxes {
			pt1 := image.Pt(zone.Min.X+box.Box.Min.X, zone.Min.Y+box.Box.Min.Y)
			pt2 := image.Pt(zone.Min.X+box.Box.Max.X, zone.Min.Y+box.Box.Max.Y)
			gocv.Rectangle(&img, image.Rect(pt1.X, pt1.Y, pt2.X, pt2.Y), color.RGBA{255, 0, 0, 255}, 2)
		}
	}

	// Save annotated image
	annotatedOut := filepath.Join("out", fmt.Sprintf("ocr_annotated_%s.png", outputName))
	_ = gocv.IMWrite(annotatedOut, img)

	return text, nil
}
