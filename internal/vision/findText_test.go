package vision_test

import (
	"image"
	"image/color"
	"testing"

	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func TestProcessImage_DebugOutput(t *testing.T) {
	// 1. Загрузка исходного изображения
	img := gocv.IMRead("references/characters_change.png", gocv.IMReadColor)
	if img.Empty() {
		t.Fatalf("Не удалось загрузить изображение: references/characters_change.png")
	}
	defer img.Close()

	// 2. Вызов функции OCR для получения результатов
	results, err := vision.ProcessImage("references/characters_change.png")
	if err != nil {
		t.Fatalf("Ошибка при вызове ProcessImage: %v", err)
	}

	// 3. Клонирование изображения для рисования
	debugImg := img.Clone()
	defer debugImg.Close()

	// 4. Настройки рисования
	green := color.RGBA{0, 255, 0, 255}
	fontFace := gocv.FontHersheyPlain
	fontScale := 1.2
	thickness := 2

	// 5. Отрисовка прямоугольников и текста
	for _, ocr := range results {
		box := image.Rect(ocr.X, ocr.Y, ocr.X+ocr.Width, ocr.Y+ocr.Height)
		gocv.Rectangle(&debugImg, box, green, thickness)

		textOrg := image.Pt(ocr.X, ocr.Y-5) // чуть выше прямоугольника
		gocv.PutText(&debugImg, ocr.Text, textOrg, fontFace, fontScale, green, 1)
	}

	// 6. Сохраняем изображение
	if ok := gocv.IMWrite("out/debug_ocr.jpg", debugImg); !ok {
		t.Errorf("Не удалось сохранить изображение в out/debug_ocr.jpg")
	}
}
