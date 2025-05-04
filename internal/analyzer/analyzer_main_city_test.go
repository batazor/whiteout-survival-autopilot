package analyzer_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func TestAnalyzeMainCityScreen(t *testing.T) {
	// Подгружаем конфиг с зонами и правилами
	areas, err := config.LoadAreaReferences("../../references/area.json")
	assert.NoError(t, err)

	rules, err := config.LoadAnalyzeRules("../../references/analyze.yaml")
	assert.NoError(t, err)

	// Настраиваем логгер (в файл)
	logPath := "../../logs/test_main_city.log"
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	logFile, err := os.Create(logPath)
	assert.NoError(t, err)
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	// Инициализируем analyzer
	an := analyzer.NewAnalyzer(areas, logger)

	// Заготовка старого состояния
	oldState := &domain.Gamer{}

	// Путь к картинке экрана
	screenshotPath := "../../references/screenshots/city_main.png"

	// Анализируем
	screen := "main_city"
	newState, err := an.AnalyzeAndUpdateState(screenshotPath, oldState, rules[screen], nil)
	assert.NoError(t, err)

	// Логи
	t.Logf("Power: %d", newState.Power)
	t.Logf("VIP Level: %d", newState.VIP.Level)
	t.Logf("Alliance Help: %v", newState.Alliance.State.IsNeedSupport)

	// Ожидаемые значения
	assert.Equal(t, 13350651, newState.Power)
	assert.Equal(t, 6, newState.VIP.Level)
	assert.True(t, newState.Alliance.State.IsNeedSupport)
}
