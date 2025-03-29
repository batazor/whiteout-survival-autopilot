package imagefinder

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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

			dominant, err := IsColorDominant(tt.imagePath, regionDef.Zone, tt.expectedColor, tt.threshold, logger)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDominant, dominant)
		})
	}
}
