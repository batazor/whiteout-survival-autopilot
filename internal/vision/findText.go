package vision

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"strings"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

type OCRTextMatch struct {
	Text       string
	Confidence float64
	Polygon    []image.Point
}

func ExtractTextContoursWithConfidence(img image.Image, region image.Rectangle) ([]OCRTextMatch, error) {
	return ExtractTextContoursWithConfidenceMultiPass(img, region)
}

func ExtractTextContoursWithConfidenceMultiPass(img image.Image, region image.Rectangle) ([]OCRTextMatch, error) {
	var all []OCRTextMatch

	passes := []func(image.Image, image.Rectangle) ([]OCRTextMatch, error){
		extractStandardPass,
		extractAggressivePass,
		extractZoomedPass,
	}

	for _, pass := range passes {
		results, err := pass(img, region)
		if err == nil {
			all = append(all, results...)
		}
	}

	return deduplicateMatches(all), nil
}

func deduplicateMatches(in []OCRTextMatch) []OCRTextMatch {
	seen := map[string]bool{}
	var out []OCRTextMatch
	for _, m := range in {
		key := fmt.Sprintf("%s@%d,%d", strings.ToLower(m.Text), m.Polygon[0].X, m.Polygon[0].Y)
		if !seen[key] {
			seen[key] = true
			out = append(out, m)
		}
	}
	return out
}

// ------------------ OCR Passes ----------------------

func extractStandardPass(img image.Image, region image.Rectangle) ([]OCRTextMatch, error) {
	return runOCRPipeline(img, region, false, func(mat gocv.Mat) gocv.Mat {
		gray := gocv.NewMat()
		gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)
		gocv.EqualizeHist(gray, &gray)
		gocv.AdaptiveThreshold(gray, &gray, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 11, 2)
		return gray
	})
}

func extractAggressivePass(img image.Image, region image.Rectangle) ([]OCRTextMatch, error) {
	return runOCRPipeline(img, region, true, func(mat gocv.Mat) gocv.Mat {
		gray := gocv.NewMat()
		gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)
		gocv.EqualizeHist(gray, &gray)
		gocv.BitwiseNot(gray, &gray)

		if gray.Cols() >= 3 && gray.Rows() >= 3 {
			gocv.Resize(gray, &gray, image.Point{}, 2.0, 2.0, gocv.InterpolationLinear)
		}

		gocv.AdaptiveThreshold(gray, &gray, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 11, 2)

		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.MorphologyEx(gray, &gray, gocv.MorphClose, kernel)
		kernel.Close()

		return gray
	})
}

func extractZoomedPass(img image.Image, region image.Rectangle) ([]OCRTextMatch, error) {
	return runOCRPipeline(img, region, true, func(mat gocv.Mat) gocv.Mat {
		gray := gocv.NewMat()
		gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)

		if gray.Cols() >= 3 && gray.Rows() >= 3 {
			gocv.Resize(gray, &gray, image.Point{}, 2.5, 2.5, gocv.InterpolationCubic)
		}

		gocv.EqualizeHist(gray, &gray)
		gocv.AdaptiveThreshold(gray, &gray, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 13, 2)
		return gray
	})
}

// ------------------ OCR Core ----------------------

func runOCRPipeline(img image.Image, region image.Rectangle, isScaled bool, preprocess func(gocv.Mat) gocv.Mat) ([]OCRTextMatch, error) {
	cropped := image.NewRGBA(region.Bounds())
	draw.Draw(cropped, region.Bounds(), img, region.Min, draw.Src)

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, cropped); err != nil {
		return nil, err
	}
	mat, err := gocv.IMDecode(buf.Bytes(), gocv.IMReadColor)
	if err != nil {
		return nil, err
	}
	defer mat.Close()

	processed := preprocess(mat)
	defer processed.Close()

	finalBuf := new(bytes.Buffer)
	finalImage, err := processed.ToImage()
	if err != nil {
		return nil, err
	}
	if err := png.Encode(finalBuf, finalImage); err != nil {
		return nil, err
	}

	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("eng")
	client.SetWhitelist("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789:.,")

	if err := client.SetImageFromBytes(finalBuf.Bytes()); err != nil {
		return nil, err
	}

	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return nil, err
	}

	var matches []OCRTextMatch
	scale := 2
	if !isScaled {
		scale = 1
	}

	for _, box := range boxes {
		if len(box.Word) < 2 || box.Confidence < 30 {
			continue
		}
		if box.Box.Dx() < 10 || box.Box.Dy() < 10 {
			continue
		}

		text := cleanText(box.Word)
		conf := float64(box.Confidence) / 100.0

		points := []image.Point{
			{X: box.Box.Min.X/scale + region.Min.X, Y: box.Box.Min.Y/scale + region.Min.Y},
			{X: box.Box.Max.X/scale + region.Min.X, Y: box.Box.Min.Y/scale + region.Min.Y},
			{X: box.Box.Max.X/scale + region.Min.X, Y: box.Box.Max.Y/scale + region.Min.Y},
			{X: box.Box.Min.X/scale + region.Min.X, Y: box.Box.Max.Y/scale + region.Min.Y},
		}

		matches = append(matches, OCRTextMatch{
			Text:       text,
			Confidence: conf,
			Polygon:    points,
		})
	}

	return matches, nil
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 {
			return r
		}
		return -1
	}, s)
	return s
}
