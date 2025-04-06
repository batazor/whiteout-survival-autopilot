package vision_test

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func cleanAndParseInt(s string) int {
	s = strings.ReplaceAll(s, " ", "")
	val, _ := strconv.Atoi(s)
	return val
}

func TestOCR_ExtractPowerAndVIP(t *testing.T) {
	// Load area.json
	lookup, err := config.LoadAreaReferences("../../references/area.json")
	if err != nil {
		t.Fatalf("failed to load area.json: %v", err)
	}

	// Prepare test data
	tests := []struct {
		name     string
		expected int
	}{
		{"power", 13350651},
		{"vip_level", 6},
		{"gems", 9727},
	}

	imagePath := filepath.Join("../../references", "screenshots", "city_main.png")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, err := lookup.GetRegionByName(tt.name)
			if err != nil {
				t.Fatalf("region not found: %v", err)
			}

			rect := region.ToRectangle()
			text, err := vision.ExtractTextFromRegion(imagePath, rect, tt.name, false)
			if err != nil {
				t.Fatalf("OCR failed: %v", err)
			}

			got := cleanAndParseInt(text)
			if got != tt.expected {
				t.Errorf("unexpected value for %s: got %d, want %d (raw: %q)", tt.name, got, tt.expected, text)
			}
		})
	}
}
