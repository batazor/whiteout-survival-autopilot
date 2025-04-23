package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

func main() {
	areaFile := "references/area.json"
	outDir := "references/icons/event/marvelous_fantasia"

	// Жесткая мапа OCR → путь к локальному скриншоту
	screenshotMap := map[string]string{
		"/data/upload/1/46441552-marvelous_fantasia_main.png":    "references/screenshots/events/marvelous_fantasia/marvelous_fantasia_main.png",
		"/data/upload/1/7f0a99ba-marvelous_fantasia_level_1.png": "references/screenshots/events/marvelous_fantasia/marvelous_fantasia_level_1.png",
	}

	// Загружаем разметку
	areaLookup, err := config.LoadAreaReferences(areaFile)
	if err != nil {
		panic("failed to load area references: " + err.Error())
	}

	// Создаём выходную директорию
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic("failed to create output directory: " + err.Error())
	}

	for _, area := range areaLookup.Areas {
		imagePath, ok := screenshotMap[area.OCR]
		if !ok {
			continue
		}

		// Загружаем скриншот
		f, err := os.Open(imagePath)
		if err != nil {
			panic("failed to open image: " + err.Error())
		}
		img, _, err := image.Decode(f)
		_ = f.Close()
		if err != nil {
			panic("failed to decode image: " + err.Error())
		}

		for idx, name := range area.Transcription {
			if !strings.HasPrefix(name, "marvelous_fantasia_icon_") || idx >= len(area.BBox) {
				continue
			}

			bbox := area.BBox[idx]
			x, y, w, h := bbox.ToPixels()
			region := image.Rect(x, y, x+w, y+h)

			subImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(region)

			// Считаем SHA1-хеш от PNG-содержимого иконки
			buf := new(bytes.Buffer)
			if err := png.Encode(buf, subImg); err != nil {
				panic("failed to encode icon to buffer: " + err.Error())
			}
			hash := sha1.Sum(buf.Bytes())
			hashStr := hex.EncodeToString(hash[:])[:5]

			// Сохраняем в файл
			fileName := name + "_" + hashStr + ".png"
			outPath := filepath.Join(outDir, fileName)

			of, err := os.Create(outPath)
			if err != nil {
				panic("failed to create output file: " + err.Error())
			}
			if _, err := of.Write(buf.Bytes()); err != nil {
				of.Close()
				panic("failed to write image file: " + err.Error())
			}
			of.Close()
		}
	}
}
