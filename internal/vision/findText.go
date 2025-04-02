package vision

import (
	"fmt"
	"image"
	"sync"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// ProcessImage applies all OCR strategies in parallel and returns the aggregated OCR results.
func ProcessImage(imagePath string) ([]domain.OCRResult, error) {
	// Read the image from file
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return nil, fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	// List of strategy functions to apply
	strategies := []func(gocv.Mat) (gocv.Mat, error){
		PreprocessForOCR,
	}

	// Channel to collect results from goroutines
	type ocrResultSet struct {
		results []domain.OCRResult
		err     error
	}
	resultsChan := make(chan ocrResultSet, len(strategies))
	var wg sync.WaitGroup

	// Launch each strategy in a separate goroutine
	for _, stratFn := range strategies {
		wg.Add(1)
		go func(fn func(gocv.Mat) (gocv.Mat, error)) {
			defer wg.Done()
			// Apply preprocessing strategy
			processedMat, err := fn(img)
			if err != nil {
				resultsChan <- ocrResultSet{nil, err}
				return
			}
			defer processedMat.Close()
			// Scale the processed image up by 2x for better OCR accuracy
			scaledMat := gocv.NewMat()
			gocv.Resize(processedMat, &scaledMat, image.Point{}, 2.0, 2.0, gocv.InterpolationNearestNeighbor)
			// Perform OCR using Tesseract (via gosseract)
			client := gosseract.NewClient()
			defer client.Close()
			// (Optional: set language if needed, e.g., client.SetLanguage("eng"))
			// Encode the scaled image to bytes (PNG format for lossless compression)
			buf, err := gocv.IMEncode(".png", scaledMat)
			scaledMat.Close()
			if err != nil {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("image encode error: %v", err)}
				return
			}
			if err := client.SetImageFromBytes(buf.GetBytes()); err != nil {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("tesseract set image error: %v", err)}
				return
			}
			boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
			if err != nil {
				resultsChan <- ocrResultSet{nil, fmt.Errorf("tesseract OCR error: %v", err)}
				return
			}
			// Filter out words with 2 or fewer characters, or confidence below 40
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

	// Wait for all goroutines to finish and close the channel
	wg.Wait()
	close(resultsChan)

	// Collect results and check for errors
	var allResults []domain.OCRResult
	var collectErr error
	for res := range resultsChan {
		if res.err != nil {
			collectErr = res.err
			// continue gathering (all goroutines have finished) to close channel properly
		}
		if res.results != nil {
			allResults = append(allResults, res.results...)
		}
	}
	if collectErr != nil {
		// If any strategy failed, return the error (no results)
		return nil, collectErr
	}

	// Remove duplicate entries (same text and overlapping coordinates)
	deduped := removeDuplicates(allResults)
	return deduped, nil
}

// removeDuplicates filters out duplicate OCR results by text and location overlap.
func removeDuplicates(results []domain.OCRResult) []domain.OCRResult {
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
				// Compute bounding boxes for both results
				rectA := image.Rect(results[i].X, results[i].Y, results[i].X+results[i].Width, results[i].Y+results[i].Height)
				rectB := image.Rect(results[j].X, results[j].Y, results[j].X+results[j].Width, results[j].Y+results[j].Height)
				// Check if bounding boxes overlap
				inter := rectA.Intersect(rectB)
				if inter.Dx() > 0 && inter.Dy() > 0 {
					// If overlapping and same text, mark the one with lower confidence as duplicate
					if results[i].Confidence >= results[j].Confidence {
						marked[j] = true
					} else {
						marked[i] = true
						break // current i is inferior, stop comparing it with others
					}
				}
			}
		}
	}
	// Build a new slice excluding marked duplicates
	deduped := make([]domain.OCRResult, 0, len(results))
	for idx, res := range results {
		if !marked[idx] {
			deduped = append(deduped, res)
		}
	}
	return deduped
}

// PreprocessForOCR обрабатывает входное изображение с белым текстом на зелёном фоне
// и возвращает Mat с белым текстом на чёрном фоне.
func PreprocessForOCR(src gocv.Mat) (gocv.Mat, error) {
	// 1. Преобразование в HSV и маскирование фона указанного зелёного цвета
	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(src, &hsv, gocv.ColorBGRToHSV)
	// Определяем диапазон HSV для светло-зелёного фона (примерно Hue ~ 60 ±10, Sat ~ 90-255, Val ~ 90-255)
	lower := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(50.0, 90.0, 90.0, 0.0), hsv.Rows(), hsv.Cols(), gocv.MatTypeCV8UC3)
	defer lower.Close()
	upper := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(70.0, 255.0, 255.0, 0.0), hsv.Rows(), hsv.Cols(), gocv.MatTypeCV8UC3)
	defer upper.Close()
	mask := gocv.NewMat()
	defer mask.Close()
	gocv.InRange(hsv, lower, upper, &mask) // теперь mask имеет белый фон и чёрный текст

	// 2. Морфологические операции для усиления маски (удаление шумов, закрытие мелких дыр)
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Erode(mask, &mask, kernel)  // слегка сужаем области фона, убирая мелкие белые пятна в тексте
	gocv.Dilate(mask, &mask, kernel) // восстанавливаем основной фон, текст остаётся вырезанным чётче

	// 3. Повышение контраста: перевод в оттенки серого и гистограммная эквализация
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(src, &gray, gocv.ColorBGRToGray)
	gocv.EqualizeHist(gray, &gray) // выравниваем гистограмму, повышая контраст между текстом и фоном

	// 4. Применение маски к серому изображению и финальная бинаризация (Otsu)
	textMask := gocv.NewMat()
	defer textMask.Close()
	gocv.BitwiseNot(mask, &textMask) // инвертируем маску: белым становится текст, фон — чёрным
	textOnly := gocv.NewMat()
	defer textOnly.Close()
	gocv.BitwiseAnd(gray, textMask, &textOnly) // оставляем на изображении только текст (фон обнулён)
	binMat := gocv.NewMat()
	defer binMat.Close()
	// Пороговое преобразование с методом Отсу для получения белого текста на чёрном фоне
	gocv.Threshold(textOnly, &binMat, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	// 5. Масштабирование результата 2x для улучшения читаемости мелкого текста
	finalMat := gocv.NewMat()
	// Используем ближайшего соседа, чтобы сохранить чёткие границы текста при масштабировании
	gocv.Resize(binMat, &finalMat, image.Point{}, 1.0, 1.0, gocv.InterpolationNearestNeighbor)
	// Возвращаем полученное бинарное изображение (1 канал: белый текст на чёрном фоне)
	return finalMat, nil
}
