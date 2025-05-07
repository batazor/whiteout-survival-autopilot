package analyzer_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func TestAnalyzeAndSaveRegion(t *testing.T) {
	// Загружаем зоны (можно пустые для начала)
	areas := &config.AreaLookup{}

	// Создаём правило с saveAsRegion
	rules := []domain.AnalyzeRule{
		{
			Name:         "alliance.tech.favorite",
			Action:       "findIcon",
			Threshold:    0.86,
			SaveAsRegion: true,
		},
	}

	// Логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	an := analyzer.NewAnalyzer(areas, logger)

	// Минимальное старое состояние
	oldState := &domain.Gamer{}

	screenshotPath := "../../references/screenshots/alliance/alliance_tech.png"

	newState, err := an.AnalyzeAndUpdateState(screenshotPath, oldState, rules, nil)
	assert.NoError(t, err)
	assert.NotNil(t, newState)

	// Проверяем, что регион сохранён
	region, found := areas.Get("alliance.tech.favorite")
	assert.True(t, found, "регион должен быть сохранён в areas")
	assert.True(t, region.Zone.Dx() > 0 && region.Zone.Dy() > 0, "размер зоны должен быть > 0")
}
