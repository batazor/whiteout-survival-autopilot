package domain

// OCRResult holds a recognized word with its confidence and bounding box.
type OCRResult struct {
	Text       string
	Confidence float64
	X          int
	Y          int
	Width      int
	Height     int
}
