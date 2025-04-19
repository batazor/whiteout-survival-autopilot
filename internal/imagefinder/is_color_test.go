package imagefinder

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

func TestIsColorDominant_Table(t *testing.T) {
	tests := []struct {
		name          string
		imagePath     string
		regionKey     string
		expectedColor string `yaml:"expectedColor,omitempty" mapstructure:"expectedColor"`
		threshold     float32
		wantDominant  bool
	}{
		{
			name:          "Claim button is active (green)",
			imagePath:     "../../references/screenshots/exploration.png",
			regionKey:     "exploration.state.isClaimActive",
			expectedColor: "green",
			threshold:     0.7,
			wantDominant:  true,
		},
		{
			name:          "Claim button is disabled (no green)",
			imagePath:     "../../references/screenshots/exploration_disable.png",
			regionKey:     "exploration.state.isClaimActive",
			expectedColor: "green",
			threshold:     0.7,
			wantDominant:  false,
		},
		{
			name:          "Gift button is available (red)",
			imagePath:     "../../references/screenshots/alliance_chest_gift.png",
			regionKey:     "alliance.state.isMainChest",
			expectedColor: "red",
			threshold:     0.7,
			wantDominant:  true,
		},
		{
			name:          "Gift button is not available (no red)",
			imagePath:     "../../references/screenshots/alliance_chest_loot.png",
			regionKey:     "alliance.state.isMainChest",
			expectedColor: "red",
			threshold:     0.7,
			wantDominant:  false,
		},
		{
			name:          "Contribute button is visible (blue)",
			imagePath:     "../../references/screenshots/alliance_tech_contribute.png",
			regionKey:     "alliance.state.isAllianceContributeButton",
			expectedColor: "blue",
			threshold:     0.5,
			wantDominant:  true,
		},
		{
			name:          "Contribute button is no visible (gray)",
			imagePath:     "../../references/screenshots/alliance_tech_contribute_disabled.png",
			regionKey:     "alliance.state.isAllianceContributeButton",
			expectedColor: "blue",
			threshold:     0.5,
			wantDominant:  false,
		},
	}

	// Загружаем конфиг с областями один раз
	areaConfig, err := config.LoadAreaReferences("../../references/area.json")
	assert.NoError(t, err)

	// Настраиваем логгер
	logPath := "../../logs/test_match_icon.log"
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	logFile, err := os.Create(logPath)
	assert.NoError(t, err)
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regionDef, ok := areaConfig.Get(tt.regionKey)
			assert.True(t, ok, "region %s should be present", tt.regionKey)

			// Обрезаем изображение
			img := gocv.IMRead(tt.imagePath, gocv.IMReadColor)
			assert.False(t, img.Empty(), "failed to load image: %s", tt.imagePath)
			defer img.Close()

			crop := img.Region(regionDef.Zone)
			defer crop.Close()

			cropPath := filepath.Join("./out", "crop_"+tt.name+".png")
			gocv.IMWrite(cropPath, crop)

			// Проверяем цвет
			dominant, _ := IsColorDominant(tt.imagePath, regionDef.Zone, tt.expectedColor, tt.threshold, logger)
			assert.Equal(t, tt.wantDominant, dominant)
		})
	}
}
