package vision

import (
	"fmt"
	"image"
	"sort"
	"sync"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// ProcessImage applies all OCR strategies in parallel and returns the aggregated OCR results.
func ProcessImage(imagePath string) (domain.OCRResults, error) {
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return nil, fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	strategies := []func(gocv.Mat) (gocv.Mat, error){
		PreprocessForOCR,
		PreprocessHLS,
		PreprocessYCrCb,
	}

	type ocrResultSet struct {
		results domain.OCRResults
		err     error
	}
	resultsChan := make(chan ocrResultSet, len(strategies))
	var wg sync.WaitGroup

	for _, stratFn := range strategies {
		wg.Add(1)
		go func(fn func(gocv.Mat) (gocv.Mat, error)) {
			defer wg.Done()

			processedMat, err := fn(img)
			if err != nil || processedMat.Empty() {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("preprocessing error: %v", err)}
				return
			}
			defer processedMat.Close()

			// Upscale x2 for better OCR accuracy
			scaled := gocv.NewMat()
			gocv.Resize(processedMat, &scaled, image.Point{}, 2.0, 2.0, gocv.InterpolationNearestNeighbor)
			defer scaled.Close()

			client := gosseract.NewClient()
			defer client.Close()

			client.SetLanguage("eng")
			client.SetPageSegMode(gosseract.PSM_AUTO)
			client.SetVariable("user_defined_dpi", "814") // 407 ppi * 2

			buf, err := gocv.IMEncode(".png", scaled)
			if err != nil {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("image encode error: %v", err)}
				return
			}
			if err := client.SetImageFromBytes(buf.GetBytes()); err != nil {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("tesseract set image error: %v", err)}
				return
			}

			boxes, err := client.GetBoundingBoxes(gosseract.RIL_TEXTLINE)
			if err != nil {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("tesseract OCR error: %v", err)}
				return
			}

			var ocrResults []domain.OCRResult
			for _, box := range boxes {
				if len(box.Word) > 2 && box.Confidence >= 10 {
					rect := box.Box
					scale := 2
					ocrResults = append(ocrResults, domain.OCRResult{
						Text:       box.Word,
						Confidence: box.Confidence,
						X:          rect.Min.X / scale,
						Y:          rect.Min.Y / scale,
						Width:      rect.Dx() / scale,
						Height:     rect.Dy() / scale,
					})
				}
			}
			resultsChan <- ocrResultSet{ocrResults, nil}
		}(stratFn)
	}

	wg.Wait()
	close(resultsChan)

	var allResults []domain.OCRResult
	var collectErr error
	for res := range resultsChan {
		if res.err != nil {
			collectErr = res.err
		}
		if res.results != nil {
			allResults = append(allResults, res.results...)
		}
	}
	if collectErr != nil {
		return nil, collectErr
	}

	// Remove duplicates and sort results
	list := removeDuplicates(allResults)
	sort.Sort(list)

	return list, nil
}

// removeDuplicates filters out duplicate OCR results by text and location overlap.
func removeDuplicates(results domain.OCRResults) domain.OCRResults {
	marked := make([]bool, len(results))
	for i := 0; i < len(results); i++ {
		if marked[i] {
			continue
		}
		for j := i + 1; j < len(results); j++ {
			if marked[j] {
				continue
			}
			if results[i].Text == results[j].Text {
				rectA := image.Rect(results[i].X, results[i].Y, results[i].X+results[i].Width, results[i].Y+results[i].Height)
				rectB := image.Rect(results[j].X, results[j].Y, results[j].X+results[j].Width, results[j].Y+results[j].Height)
				inter := rectA.Intersect(rectB)
				if inter.Dx() > 0 && inter.Dy() > 0 {
					if results[i].Confidence >= results[j].Confidence {
						marked[j] = true
					} else {
						marked[i] = true
						break
					}
				}
			}
		}
	}
	deduped := make([]domain.OCRResult, 0, len(results))
	for idx, res := range results {
		if !marked[idx] {
			deduped = append(deduped, res)
		}
	}
	return deduped
}

// PreprocessForOCR усиливает контраст и превращает картинку в бинарную
func PreprocessForOCR(src gocv.Mat) (gocv.Mat, error) {
	lab := gocv.NewMat()
	gocv.CvtColor(src, &lab, gocv.ColorBGRToLab)
	defer lab.Close()

	ch := gocv.Split(lab)
	if len(ch) < 1 {
		return gocv.NewMat(), fmt.Errorf("split returned no channels")
	}
	L := ch[0]
	defer func() {
		for _, c := range ch {
			c.Close()
		}
	}()
	clahe := gocv.NewCLAHEWithParams(2.0, image.Pt(8, 8))
	defer clahe.Close()
	clahe.Apply(L, &L)

	bin := gocv.NewMat()
	gocv.AdaptiveThreshold(L, &bin, 255,
		gocv.AdaptiveThresholdMean, gocv.ThresholdBinaryInv, 31, 15)
	defer bin.Close()

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.MorphologyEx(bin, &bin, gocv.MorphClose, kernel)

	return bin.Clone(), nil
}

func PreprocessYCrCb(src gocv.Mat) (gocv.Mat, error) {
	ycc := gocv.NewMat()
	defer ycc.Close()
	gocv.CvtColor(src, &ycc, gocv.ColorBGRToYCrCb)

	channels := gocv.Split(ycc)
	y := channels[0]

	smooth := gocv.NewMat()
	defer smooth.Close()
	gocv.BilateralFilter(y, &smooth, 9, 75, 75)

	bin := gocv.NewMat()
	defer bin.Close()
	gocv.Threshold(smooth, &bin, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	return bin.Clone(), nil
}

func PreprocessHLS(src gocv.Mat) (gocv.Mat, error) {
	hls := gocv.NewMat()
	defer hls.Close()
	gocv.CvtColor(src, &hls, gocv.ColorBGRToHLSFull)

	lChan := gocv.Split(hls)[1]

	gocv.EqualizeHist(lChan, &lChan)
	bin := gocv.NewMat()
	defer bin.Close()
	gocv.AdaptiveThreshold(lChan, &bin, 255,
		gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinaryInv, 25, 10)

	return bin.Clone(), nil
}
