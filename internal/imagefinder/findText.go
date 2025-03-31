package imagefinder

import (
	"image"
	"strings"

	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func FindTextContours(img image.Image, target string, threshold float64) ([]vision.OCRTextMatch, error) {
	region := image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy())
	all, err := vision.ExtractTextContoursWithConfidence(img, region)
	if err != nil {
		return nil, err
	}

	var results []vision.OCRTextMatch
	for _, entry := range all {
		if entry.Confidence >= threshold && strings.EqualFold(entry.Text, target) {
			results = append(results, entry)
		}
	}
	if len(results) == 0 {
		return nil, imagefinder.ErrNotFound
	}
	return results, nil
}
