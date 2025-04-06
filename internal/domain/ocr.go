package domain

type OCRResults []OCRResult

// OCRResult holds a recognized word with its confidence and bounding box.
type OCRResult struct {
	Text       string
	Confidence float64
	X          int
	Y          int
	Width      int
	Height     int
}

// Len returns the number of OCR results.
func (r OCRResults) Len() int {
	return len(r)
}

// Less sorts by confidence (descending).
func (r OCRResults) Less(i, j int) bool {
	return r[i].Confidence > r[j].Confidence
}

// Swap swaps two elements in the slice.
func (r OCRResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
