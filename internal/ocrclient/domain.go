package ocrclient

import (
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// Region describes a rectangular area in pixel coordinates.
type Region struct {
	X0 int `json:"x0"`
	Y0 int `json:"y0"`
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
}

// OCRZone is one OCR result block.
type OCRZone struct {
	Box      [][]int `json:"box"`
	Text     string  `json:"text"`
	Score    float64 `json:"score"`
	AvgColor string  `json:"avg_color"`
	BgColor  string  `json:"bg_color"`
}

// ToOCRResult converts an OCRZone (with arbitrary polygonal Box) into a rectangular OCRResult.
func (z OCRZone) ToOCRResult() domain.OCRResult {
	// Initialize min/max using the first point, if any.
	var minX, minY, maxX, maxY int
	if len(z.Box) > 0 && len(z.Box[0]) == 2 {
		minX, minY = z.Box[0][0], z.Box[0][1]
		maxX, maxY = minX, minY
	}
	// Expand the bounds to include all points in Box
	for _, pt := range z.Box {
		x, y := pt[0], pt[1]
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	return domain.OCRResult{
		Text:     z.Text,
		Score:    z.Score,
		X:        minX,
		Y:        minY,
		Width:    maxX - minX,
		Height:   maxY - minY,
		AvgColor: z.AvgColor,
		BgColor:  z.BgColor,
	}
}
