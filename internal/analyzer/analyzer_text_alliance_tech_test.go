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

func TestAnalyzeAndSaveRegion_FindText(t *testing.T) {
	// Подготовка зоны
	areas := &config.AreaLookup{}

	// Правило поиска текста с сохранением зоны
	rules := []domain.AnalyzeRule{
		{
			Name:         "alliance.state.isAllianceTechButton",
			Action:       "findText",
			Text:         "Tech",
			Threshold:    0.4,
			SaveAsRegion: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	an := analyzer.NewAnalyzer(areas, logger)

	oldState := &domain.Gamer{}
	screenshotPath := "../../references/screenshots/alliance_tech.png"

	newState, err := an.AnalyzeAndUpdateState(screenshotPath, oldState, rules, nil)
	assert.NoError(t, err)
	assert.NotNil(t, newState)

	region, found := areas.Get("alliance.state.isAllianceTechButton")
	assert.True(t, found, "регион должен быть сохранён в areas")
	assert.True(t, region.Zone.Dx() > 0 && region.Zone.Dy() > 0, "размер зоны должен быть > 0")
}
