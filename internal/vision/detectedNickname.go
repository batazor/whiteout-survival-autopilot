package vision

import (
	"fmt"
	"image"
	"strings"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

// ExtractNicknameFromImage извлекает и распознаёт никнейм из изображения по заданной зоне.
func ExtractNicknameFromImage(imagePath string, zone image.Rectangle) (string, error) {
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return "", fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	if zone.Dx() <= 0 || zone.Dy() <= 0 {
		return "", fmt.Errorf("invalid nickname region: %v", zone)
	}

	sub := img.Region(zone)
	mat := sub.Clone()
	sub.Close()

	defer mat.Close()
	pre, err := PreprocessForNicknames(mat)
	if err != nil {
		return "", fmt.Errorf("nickname preprocessing failed: %w", err)
	}
	defer pre.Close()

	scaled := gocv.NewMat()
	gocv.Resize(pre, &scaled, image.Point{}, 2.0, 2.0, gocv.InterpolationNearestNeighbor)
	defer scaled.Close()

	client := gosseract.NewClient()
	defer client.Close()

	client.SetLanguage("eng")
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)
	client.SetVariable("tessedit_char_whitelist", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789[]_-")
	client.SetVariable("user_defined_dpi", "1200")

	buf, err := gocv.IMEncode(".png", scaled)
	if err != nil {
		return "", fmt.Errorf("encode failed: %w", err)
	}
	if err := client.SetImageFromBytes(buf.GetBytes()); err != nil {
		return "", fmt.Errorf("set image failed: %w", err)
	}

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("OCR failed: %w", err)
	}

	return strings.ToLower(strings.TrimSpace(text)), nil
}

func PreprocessForNicknames(src gocv.Mat) (gocv.Mat, error) {
	// Перевод в оттенки серого
	gray := gocv.NewMat()
	gocv.CvtColor(src, &gray, gocv.ColorBGRToGray)
	defer gray.Close()

	// Применяем CLAHE для локального контраста
	clahe := gocv.NewCLAHEWithParams(3.0, image.Pt(8, 8))
	defer clahe.Close()
	claheResult := gocv.NewMat()
	clahe.Apply(gray, &claheResult)
	defer claheResult.Close()

	// Жесткий порог — для символов это чаще лучше чем adaptive
	thresh := gocv.NewMat()
	gocv.Threshold(claheResult, &thresh, 0, 255, gocv.ThresholdBinary+gocv.ThresholdOtsu)
	defer thresh.Close()

	// Морфологическое закрытие — убрать дырки и шум
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(2, 2))
	defer kernel.Close()
	closed := gocv.NewMat()
	gocv.MorphologyEx(thresh, &closed, gocv.MorphClose, kernel)

	// Увеличим резкость (high-pass фильтр)
	sharp := gocv.NewMat()
	gocv.AddWeighted(claheResult, 1.5, closed, -0.5, 0, &sharp)

	return sharp.Clone(), nil
}
