package vision

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
)

// WaitForText ищет указанный текст на экране с интервалом.
// region — область поиска (если пустая, берется вся картинка).
func WaitForText(
	ctx context.Context,
	adb adb.DeviceController,
	targetTexts []string,
	interval time.Duration,
	region image.Rectangle,
) (*OCRResult, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for text: %w", ctx.Err())

		case <-ticker.C:
			path := filepath.Join("out", fmt.Sprintf("waitfor_text_%d.png", attempt))
			img, err := adb.Screenshot(path)
			if err != nil {
				continue
			}

			if region.Empty() {
				region = img.Bounds()
			}
			cropped := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(region)

			buf := new(bytes.Buffer)
			if err := png.Encode(buf, cropped); err != nil {
				continue
			}
			mat, err := gocv.IMDecode(buf.Bytes(), gocv.IMReadColor)
			if err != nil {
				continue
			}
			defer mat.Close()

			results, err := ProcessImageFromMat(mat)
			if err != nil {
				continue
			}

			for _, match := range results {
				if match.Confidence < 10 {
					continue
				}
				text := strings.ToLower(match.Text)
				for _, target := range targetTexts {
					if strings.Contains(text, strings.ToLower(target)) {
						return &match, nil
					}
				}
			}
		}
	}
}

// ProcessImageFromMat обёртка вокруг ProcessImage, но принимает Mat напрямую
func ProcessImageFromMat(mat gocv.Mat) ([]OCRResult, error) {
	// Временный файл
	tempPath := filepath.Join("/tmp", fmt.Sprintf("mat_ocr_%d.png", time.Now().UnixNano()))
	if ok := gocv.IMWrite(tempPath, mat); !ok {
		return nil, fmt.Errorf("failed to write temp file")
	}
	defer func() { _ = os.Remove(tempPath) }()
	return ProcessImage(tempPath)
}
