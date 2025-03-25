package imagefinder

import (
	"fmt"
	"image"
	"math"
	"strings"

	"gocv.io/x/gocv"
)

// IsColorDominant проверяет, доминирует ли ожидаемый цвет в указанном регионе изображения.
func IsColorDominant(imagePath string, region image.Rectangle, expected string, threshold float32) (bool, error) {
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return false, fmt.Errorf("failed to load image: %s", imagePath)
	}
	defer img.Close()

	crop := img.Region(region)
	defer crop.Close()

	// Используем встроенный метод Mean()
	mean := crop.Mean()

	// BGR порядок
	blue := mean.Val1
	green := mean.Val2
	red := mean.Val3

	switch strings.ToLower(expected) {
	case "green":
		if green > red+30 && green > blue+30 && green > float64(threshold*255) {
			return true, nil
		}
	case "gray":
		diff1 := math.Abs(blue - green)
		diff2 := math.Abs(green - red)
		diff3 := math.Abs(blue - red)
		if diff1 < 15 && diff2 < 15 && diff3 < 15 && blue < 200 {
			return true, nil
		}
	default:
		return false, fmt.Errorf("unsupported expected color: %s", expected)
	}

	return false, nil
}
