package vision_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func TestOCR_ExtractBattleStatus(t *testing.T) {
	// Load area.json
	lookup, err := config.LoadAreaReferences("../../references/area.json")
	require.NoError(t, err, "failed to load area.json")

	regionName := "exploration.state.battleStatus"
	region, err := lookup.GetRegionByName(regionName)
	require.NoError(t, err, "region not found")

	// Define test cases
	tests := []struct {
		imageFile string
		expected  string
	}{
		{"exploration_battle_defeat.png", "Defeat"},
		{"exploration_battle_defeat_with_ads.png", "Defeat"},
		{"exploration_battle_viktory.png", "Victory"},
	}

	for _, tt := range tests {
		t.Run(tt.imageFile, func(t *testing.T) {
			imagePath := filepath.Join("../../references", "screenshots", tt.imageFile)
			text, err := vision.ExtractTextFromRegion(imagePath, region.ToRectangle(), regionName, true)
			require.NoError(t, err, "OCR failed")

			require.Equal(t, tt.expected, text, "OCR result does not match expected value")
		})
	}
}
