package config

import (
	"encoding/json"
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type Region struct {
	Zone image.Rectangle
}

type AreaLookup struct {
	Areas []domain.AreaReference
}

func LoadAreaReferences(file string) (*AreaLookup, error) {
	f, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("failed to read area file: %w", err)
	}

	var refs []domain.AreaReference
	if err := json.Unmarshal(f, &refs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal area json: %w", err)
	}

	return &AreaLookup{Areas: refs}, nil
}

// GetRegionByName returns a region (BBox) for a given transcription name.
func (a *AreaLookup) GetRegionByName(name string) (*domain.BBox, error) {
	for _, area := range a.Areas {
		for i, label := range area.Transcription {
			if label == name {
				if i < len(area.BBox) {
					return &area.BBox[i], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("region with name '%s' not found", name)
}

// Get returns Region for a given transcription name
func (a *AreaLookup) Get(name string) (Region, bool) {
	for _, area := range a.Areas {
		for i, label := range area.Transcription {
			if label == name && i < len(area.BBox) {
				bbox := area.BBox[i]
				x, y, w, h := bbox.ToPixels()
				return Region{
					Zone: image.Rect(x, y, x+w, y+h),
				}, true
			}
		}
	}
	return Region{}, false
}

func (a *AreaLookup) AddTemporaryRegion(name string, region Region) {
	bbox := domain.NewBBoxFromRect(region.Zone, 1080, 2400)

	a.Areas = append(a.Areas, domain.AreaReference{
		OCR:           "generated",
		ID:            -1,
		BBox:          []domain.BBox{bbox},
		Transcription: []string{name},
	})

	slog.Info("🗺️ Добавлена временная зона",
		slog.String("name", name),
		slog.Float64("x", bbox.X),
		slog.Float64("y", bbox.Y),
		slog.Float64("width", bbox.Width),
		slog.Float64("height", bbox.Height),
	)
}
