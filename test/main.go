package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"gocv.io/x/gocv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: drawrect <input-image> <output-image>")
		return
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Загрузка изображения
	img := gocv.IMRead(inputPath, gocv.IMReadColor)
	if img.Empty() {
		fmt.Fprintf(os.Stderr, "Error: failed to read image from %s\n", inputPath)
		return
	}
	defer img.Close()

	// Координаты и размеры из OCR
	x, y, w, h := 255, 1237, 525, 35

	// Прямоугольник: левый-верхний и правый-нижний углы
	rect := image.Rect(x, y, x+w, y+h)

	// Рисуем границу прямоугольника (здесь зелёный цвет, толщина 2px)
	gocv.Rectangle(&img, rect, color.RGBA{G: 255, A: 0xff}, 2)

	// Сохраняем результат
	if ok := gocv.IMWrite(outputPath, img); !ok {
		fmt.Fprintf(os.Stderr, "Error: failed to write image to %s\n", outputPath)
		return
	}

	fmt.Printf("Drew rectangle at %v and saved to %s\n", rect, outputPath)
}
