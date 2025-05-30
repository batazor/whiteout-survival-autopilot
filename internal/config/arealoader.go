package config

import (
	"encoding/json"
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type Region struct {
	Zone image.Rectangle
}

type AreaLookup struct {
	refs atomic.Value
}

// LoadAreaReferences reads the file and seeds the atomic.Value
func LoadAreaReferences(file string) (*AreaLookup, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("failed to read area file: %w", err)
	}

	var initial []domain.AreaReference
	if err := json.Unmarshal(data, &initial); err != nil {
		return nil, fmt.Errorf("failed to unmarshal area json: %w", err)
	}

	al := &AreaLookup{}
	// Store a copy (just in case)
	cp := make([]domain.AreaReference, len(initial))
	copy(cp, initial)
	al.refs.Store(cp)

	return al, nil
}

// snapshot returns the current slice snapshot
func (a *AreaLookup) snapshot() []domain.AreaReference {
	return a.refs.Load().([]domain.AreaReference)
}

// Get returns Region for a given transcription name (lock-free)
func (a *AreaLookup) Get(name string) (Region, bool) {
	// load a consistent snapshot of your []AreaReference
	for _, area := range a.refs.Load().([]domain.AreaReference) {
		for i, label := range area.Transcription {
			if label == name && i < len(area.BBox) {
				b := area.BBox[i]
				x, y, w, h := b.ToPixels()
				return Region{Zone: image.Rect(x, y, x+w, y+h)}, true
			}
		}
	}
	return Region{}, false
}

// GetRegionByName does an atomic Load and is lock-free
func (a *AreaLookup) GetRegionByName(name string) (*domain.BBox, error) {
	for _, area := range a.snapshot() {
		for i, label := range area.Transcription {
			if label == name && i < len(area.BBox) {
				return &area.BBox[i], nil
			}
		}
	}
	return nil, fmt.Errorf("region '%s' not found", name)
}

// AddTemporaryRegion does copy-on-write and then a single atomic Store
func (a *AreaLookup) AddTemporaryRegion(name string, region Region) {
	old := a.snapshot()
	// make a new slice so we don't mutate the old one in-place
	newSlice := make([]domain.AreaReference, len(old))
	copy(newSlice, old)

	bbox := domain.NewBBoxFromRect(region.Zone, 1080, 2400)
	updated := false

	// try to update an existing entry
	for i := range newSlice {
		for j, label := range newSlice[i].Transcription {
			if label == name {
				if j < len(newSlice[i].BBox) {
					newSlice[i].BBox[j] = bbox
					updated = true
					slog.Info("🛠️ Updated temporary region", slog.String("name", name), slog.Float64("x", bbox.X), slog.Float64("y", bbox.Y))
				}
				break
			}
		}
		if updated {
			break
		}
	}

	// if we didn’t find it, append
	if !updated {
		newSlice = append(newSlice, domain.AreaReference{
			OCR:           "generated",
			ID:            -1,
			BBox:          []domain.BBox{bbox},
			Transcription: []string{name},
		})
		slog.Info("🗺️ Added temporary zone", slog.String("name", name), slog.Float64("x", bbox.X), slog.Float64("y", bbox.Y))
	}

	// one atomic swap
	a.refs.Store(newSlice)
}
