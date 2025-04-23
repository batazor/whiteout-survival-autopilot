package finder

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// DrawBoxes рисует найденные области поверх изображения и сохраняет debug-картинку в папку out/.
func DrawBoxes(screenshotPath string, boxes []domain.BBox) error {
	imgFile, err := os.Open(screenshotPath)
	if err != nil {
		return err
	}
	defer imgFile.Close()

	srcImg, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}

	outImg := image.NewRGBA(srcImg.Bounds())
	draw.Draw(outImg, outImg.Bounds(), srcImg, image.Point{}, draw.Src)

	green := color.RGBA{0, 255, 0, 255}
	for _, bbox := range boxes {
		x, y, w, h := bbox.ToPixels()
		drawRect(outImg, image.Rect(x, y, x+w, y+h), green)
	}

	// Создаём выходной путь
	if err := os.MkdirAll("out", 0755); err != nil {
		return err
	}
	base := strings.TrimSuffix(filepath.Base(screenshotPath), filepath.Ext(screenshotPath))
	outPath := filepath.Join("out", base+"_debug.png")

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, outImg)
}

func drawRect(img *image.RGBA, rect image.Rectangle, col color.Color) {
	thickness := 5

	// Вверх и низ
	for dy := 0; dy < thickness; dy++ {
		for x := rect.Min.X; x <= rect.Max.X; x++ {
			img.Set(x, rect.Min.Y+dy, col) // верх
			img.Set(x, rect.Max.Y-dy, col) // низ
		}
	}

	// Лево и право
	for dx := 0; dx < thickness; dx++ {
		for y := rect.Min.Y; y <= rect.Max.Y; y++ {
			img.Set(rect.Min.X+dx, y, col) // лево
			img.Set(rect.Max.X-dx, y, col) // право
		}
	}
}
