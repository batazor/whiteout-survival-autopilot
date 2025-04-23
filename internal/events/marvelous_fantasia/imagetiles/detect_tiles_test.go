package imagetiles_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	finder "github.com/batazor/whiteout-survival-autopilot/internal/analyzer/findIcon"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/events/marvelous_fantasia/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/events/marvelous_fantasia/imagetiles"
)

func TestDetectTiles(t *testing.T) {
	areaLookup, err := config.LoadAreaReferences("../../../../references/area.json")
	require.NoError(t, err, "не удалось загрузить area.json")

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	tests := []struct {
		name       string
		screenshot string
		wantTiles  []domain.Tile
	}{
		{
			name:       "level1",
			screenshot: "../../../../references/screenshots/events/marvelous_fantasia/marvelous_fantasia_level_1.png",
			wantTiles: []domain.Tile{
				{Kind: "star", Level: 1},
				{Kind: "star", Level: 1},
				{Kind: "star", Level: 1},
				{Kind: "star", Level: 1},
				{Kind: "wrench", Level: 1},
				{Kind: "wrench", Level: 1},
				{Kind: "wrench", Level: 1},
				{Kind: "wrench", Level: 1},
				{Kind: "bear", Level: 1},
				{Kind: "bear", Level: 2},
				{Kind: "bear", Level: 2},
				{Kind: "bear", Level: 2},
				{Kind: "bear", Level: 2},
				{Kind: "bear", Level: 2},
				{Kind: "wrench", Level: 2},
				{Kind: "wrench", Level: 2},
				{Kind: "star", Level: 2},
				{Kind: "star", Level: 2},
				// можно дописать остальные
			},
		},
	}

	iconDir := "../../../../references/icons/event/marvelous_fantasia"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			tiles, err := imagetiles.DetectTiles(tt.screenshot, areaLookup, logger, iconDir)
			require.NoError(err, "DetectTiles вернул ошибку")

			assert.Equal(t, len(tt.wantTiles), len(tiles), "не совпадает количество плиток")

			// Отрисовываем через универсальную функцию
			err = finder.DrawBoxes(tt.screenshot, tiles.BBoxes())
			require.NoError(err, "ошибка отрисовки")
		})
	}
}
