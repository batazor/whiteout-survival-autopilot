package domain

type AreaReference struct {
	OCR           string   `json:"ocr"`
	ID            int      `json:"id"`
	BBox          []BBox   `json:"bbox"`
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
}

func (b *BBox) ToPixels() (x, y, w, h int) {
	x = int(b.X * float64(b.OriginalWidth) / 100)
	y = int(b.Y * float64(b.OriginalHeight) / 100)
	w = int(b.Width * float64(b.OriginalWidth) / 100)
	h = int(b.Height * float64(b.OriginalHeight) / 100)
	return
}
