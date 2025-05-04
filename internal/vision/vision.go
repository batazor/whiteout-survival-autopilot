package vision

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// ExtractTextFromRegion выполняет OCR для заданной зоны и подсвечивает найденные словоформы.
// Она использует vision.ProcessImage (CLAHE/HLS/YCrCb + Tesseract) и выбирает
// лучший NUMERIC-результат по максимальной Confidence.
func ExtractTextFromRegion(imagePath string, zone image.Rectangle, outputName string) (string, error) {
	// ── 1. Загружаем полный скрин ────────────────────────────────────────────────
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return "", fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	// ── 2. Вырезаем область интереса и сохраняем временно ────────────────────────
	cropped := img.Region(zone)
	defer cropped.Close()

	tmpPath := filepath.Join("out", fmt.Sprintf("ocr_crop_%s.png", outputName))
	if !gocv.IMWrite(tmpPath, cropped) {
		return "", fmt.Errorf("failed to write cropped image")
	}

	// ── 3. Запускаем многостратегийный OCR  ──────────────────────────────────────
	results, err := ProcessImage(tmpPath)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no OCR results")
	}

	// ── 4. Оставляем только числовые строки и берём самую уверенную ─────────────
	var numeric domain.OCRResults
	reDigits := regexp.MustCompile(`^\d+$`)
	for _, r := range results {
		text := strings.TrimSpace(r.Text)
		if reDigits.MatchString(text) {
			numeric = append(numeric, r)
		}
	}
	if len(numeric) == 0 {
		// fallback: вернём максимум по confidence из всех
		numeric = results
	}

	sort.SliceStable(numeric, func(i, j int) bool {
		return numeric[i].Confidence > numeric[j].Confidence
	})
	best := numeric[0]

	// ── 5. Подсвечиваем bounding boxes на исходном скрине ───────────────────────
	for _, box := range results {
		pt1 := image.Pt(zone.Min.X+box.X, zone.Min.Y+box.Y)
		pt2 := image.Pt(pt1.X+box.Width, pt1.Y+box.Height)
		gocv.Rectangle(&img, image.Rect(pt1.X, pt1.Y, pt2.X, pt2.Y),
			color.RGBA{255, 0, 0, 255}, 2)
	}

	annotatedOut := filepath.Join("out", fmt.Sprintf("ocr_annotated_%s.png", outputName))
	_ = gocv.IMWrite(annotatedOut, img)

	return strings.TrimSpace(best.Text), nil
}
