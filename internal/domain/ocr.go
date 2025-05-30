package domain

import (
	"fmt"
	"image"
)

type OCRResults []OCRResult

// OCRResult holds a recognized word with its confidence and bounding box.
type OCRResult struct {
	Text     string
	Score    float64
	X        int
	Y        int
	Width    int
	Height   int
	AvgColor string `json:"avg_color"`
	BgColor  string `json:"bg_color"`
}

// FilterByRect возвращает только те OCRResult, чьи боксы полностью лежат внутри заданного прямоугольника.
func (rs OCRResults) FilterByRect(rect image.Rectangle) OCRResults {
	var out OCRResults
	for _, r := range rs {
		// строим прямоугольник результата
		rRect := image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height)
		// проверяем полное вхождение
		if rect.Min.X <= rRect.Min.X &&
			rRect.Max.X <= rect.Max.X &&
			rect.Min.Y <= rRect.Min.Y &&
			rRect.Max.Y <= rect.Max.Y {
			out = append(out, r)
		}
	}
	return out
}

// FilterByBBox возвращает только те OCRResult, чьи прямоугольники лежат в BBox с допуском ±5%,
// но не менее 10px и не более 50px по каждому измерению.
func (rs OCRResults) FilterByBBox(b *BBox) OCRResults {
	rect := b.ToRectangle()
	dx := rect.Dx()
	dy := rect.Dy()

	// Вычисляем погрешность ±5% и ограничиваем [10, 50]
	marginX := int(float64(dx) * 0.05)
	if marginX < 10 {
		marginX = 10
	} else if marginX > 50 {
		marginX = 50
	}
	marginY := int(float64(dy) * 0.05)
	if marginY < 10 {
		marginY = 10
	} else if marginY > 50 {
		marginY = 50
	}

	// Расширяем исходный прямоугольник с учётом погрешности
	expRect := image.Rect(
		rect.Min.X-marginX,
		rect.Min.Y-marginY,
		rect.Max.X+marginX,
		rect.Max.Y+marginY,
	)

	var out OCRResults
	for _, r := range rs {
		rRect := image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height)
		if rRect.Overlaps(expRect) {
			out = append(out, r)
		}
	}
	return out
}

// Len returns the number of OCR results.
func (r OCRResults) Len() int {
	return len(r)
}

// Less sorts by confidence (descending).
func (r OCRResults) Less(i, j int) bool {
	return r[i].Score > r[j].Score
}

// Swap swaps two elements in the slice.
func (r OCRResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (o OCRResult) String() string {
	return fmt.Sprintf("OCRResult{text: %q, conf: %.2f, box: (%d,%d)-(%d,%d)}",
		o.Text,
		o.Score,
		o.X, o.Y,
		o.X+o.Width, o.Y+o.Height,
	)
}
