package imagefinder

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchIconInRegion(t *testing.T) {
	screenshotPath := "../../references/screenshots/city_main.png"
	iconPath := "../../references/icons/allience_help.png"
	region := Region{X: 100, Y: 100, Width: 300, Height: 300}
	threshold := float32(0.4)

	// Настроим логгер
	logFile, err := os.Create("../../logs/test_match_icon.log")
	assert.NoError(t, err)
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	// Вызов
	match, confidence, err := MatchIconInRegion(screenshotPath, iconPath, region, threshold, logger)

	// Проверки
	assert.NoError(t, err)
	t.Logf("Confidence: %.3f", confidence)
	assert.True(t, confidence >= float32(threshold), "icon should match with sufficient confidence")
	assert.True(t, match, "icon should match the region")
}
