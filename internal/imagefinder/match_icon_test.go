package imagefinder

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

func TestMatchIconInRegion(t *testing.T) {
	// Загружаем конфиг с областями
	areaConfig, err := config.LoadAreaReferences("../../references/area.json")
	assert.NoError(t, err)

	// Получаем прямоугольник для иконки
	region, ok := areaConfig.Get("allience_help")
	assert.True(t, ok)

	// Пути к изображениям
	screenshotPath := "../../references/screenshots/city_main.png"
	iconPath := "../../references/icons/allience_help.png"

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

	// Запускаем тест
	match, confidence, err := MatchIconInRegion(screenshotPath, iconPath, region.Zone, 0.4, logger)

	// Проверки
	assert.NoError(t, err)
	t.Logf("Confidence: %.3f", confidence)
	assert.True(t, confidence >= 0.4, "confidence should be above threshold")
	assert.True(t, match, "icon should match the region")
}
