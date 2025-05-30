package domain

import (
	"image"
)

type AreaReference struct {
	OCR           string   `json:"ocr"`
	ID            int      `json:"id"`
	BBox          BBoxes   `json:"bbox"`
	Transcription []string `json:"transcription"`
}

type BBox struct {
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	Rotation       float64 `json:"rotation"`
	OriginalWidth  int     `json:"original_width"`
	OriginalHeight int     `json:"original_height"`

	Confidence float32 `json:"confidence,omitempty"`
}

func (b *BBox) ToPixels() (x, y, w, h int) {
	x = int(b.X * float64(b.OriginalWidth) / 100)
	y = int(b.Y * float64(b.OriginalHeight) / 100)
	w = int(b.Width * float64(b.OriginalWidth) / 100)
	h = int(b.Height * float64(b.OriginalHeight) / 100)
	return
}

func (b *BBox) ToRectangle() image.Rectangle {
	x, y, w, h := b.ToPixels()
	return image.Rect(x, y, x+w, y+h)
}

func NewBBoxFromRect(r image.Rectangle, originalW, originalH int) BBox {
	return BBox{
		X:              float64(r.Min.X) / float64(originalW) * 100,
		Y:              float64(r.Min.Y) / float64(originalH) * 100,
		Width:          float64(r.Dx()) / float64(originalW) * 100,
		Height:         float64(r.Dy()) / float64(originalH) * 100,
		Rotation:       0,
		OriginalWidth:  originalW,
		OriginalHeight: originalH,
	}
}

type BBoxes []BBox

// GetBest возвращает BBox с максимальным Score.
func (bxs BBoxes) GetBest() (BBox, bool) {
	if len(bxs) == 0 {
		return BBox{}, false
	}
	best := bxs[0]
	for _, b := range bxs[1:] {
		if b.Confidence > best.Confidence {
			best = b
		}
	}
	return best, true
}

// GetTopY возвращает BBox с минимальной Y (расположен выше всех).
func (bxs BBoxes) GetTopY() (BBox, bool) {
	if len(bxs) == 0 {
		return BBox{}, false
	}
	top := bxs[0]
	for _, b := range bxs[1:] {
		if b.Y < top.Y {
			top = b
		}
	}
	return top, true
}
