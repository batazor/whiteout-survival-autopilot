package vision_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func TestExtractTextContoursWithConfidence(t *testing.T) {
	// Загружаем тестовое изображение
	f, err := os.Open("references/welcome_back.png")
	assert.NoError(t, err)
	defer f.Close()

	img, err := png.Decode(f)
	assert.NoError(t, err)

	// Вызываем функцию
	matches, err := vision.ExtractTextContoursWithConfidence(img, image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	assert.NoError(t, err)

	found := map[string]int{}

	for _, m := range matches {
		switch m.Text {
		case "Confirm", "Offline", "Time":
			found[m.Text]++
		}
	}

	assert.Equal(t, 1, found["Confirm"], "should find 1 'Confirm'")
	assert.Equal(t, 1, found["Time"], "should find 1 'Time'")
	assert.Equal(t, 2, found["Offline"], "should find 2 'Offline'")

	// Сохраняем debug-изображение
	err = saveOCRDebugImage(img, matches, "out/debug_ocr.jpg")
	assert.NoError(t, err)
}

func saveOCRDebugImage(img image.Image, matches []vision.OCRTextMatch, outPath string) error {
	// Преобразуем в gocv.Mat
	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)
	mat, err := gocv.IMDecode(buf.Bytes(), gocv.IMReadColor)
	if err != nil {
		return err
	}
	defer mat.Close()

	green := color.RGBA{0, 255, 0, 255}
	for _, match := range matches {
		// Рисуем полигон
		for i := 0; i < len(match.Polygon); i++ {
			p1 := match.Polygon[i]
			p2 := match.Polygon[(i+1)%len(match.Polygon)]
			gocv.Line(&mat, p1, p2, green, 2)
		}

		// Добавляем подпись
		if len(match.Text) > 0 {
			textPos := match.Polygon[0]
			gocv.PutText(&mat, match.Text, textPos, gocv.FontHersheyPlain, 1.2, green, 1)
		}
	}

	// Сохраняем результат
	_ = os.MkdirAll(filepath.Dir(outPath), 0755)
	outImg, err := mat.ToImage()
	if err != nil {
		return err
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return jpeg.Encode(outFile, outImg, &jpeg.Options{Quality: 90})
}
