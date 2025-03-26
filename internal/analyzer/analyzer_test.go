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
	an := analyzer.NewAnalyzer(areas, rules, logger)

	// Заготовка старого состояния
	oldState := &domain.State{
		Accounts: []domain.Account{
			{
				Characters: []domain.Gamer{
					{},
				},
			},
		},
	}

	// Путь к картинке экрана
	screenshotPath := "../../references/screenshots/city_main.png"

	// Анализируем
	newState, err := an.AnalyzeAndUpdateState(screenshotPath, oldState, "main_city")
	assert.NoError(t, err)

	char := newState.Accounts[0].Characters[0]

	// Логи
	t.Logf("Power: %d", char.Power)
	t.Logf("VIP Level: %d", char.VIPLevel)
	t.Logf("Alliance Help: %v", char.Alliance.State.IsNeedSupport)

	// Ожидаемые значения
	assert.Equal(t, 13350651, char.Power)
	assert.Equal(t, 6, char.VIPLevel)
	assert.True(t, char.Alliance.State.IsNeedSupport)
}
