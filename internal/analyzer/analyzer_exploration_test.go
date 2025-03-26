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

func TestAnalyzeExplorationScreens(t *testing.T) {
	areas, err := config.LoadAreaReferences("../../references/area.json")
	assert.NoError(t, err)

	rules, err := config.LoadAnalyzeRules("../../references/analyze.yaml")
	assert.NoError(t, err)

	// Логгер
	logPath := "../../logs/test_exploration_group.log"
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

	tests := []struct {
		name             string
		screenshot       string
		wantClaimActive  bool
		wantExploreLevel int // 0 means "don't check exact value"
	}{
		{
			name:             "Exploration screen - claim active",
			screenshot:       "../../references/screenshots/exploration.png",
			wantClaimActive:  true,
			wantExploreLevel: 306,
		},
		{
			name:             "Exploration screen - claim disabled",
			screenshot:       "../../references/screenshots/exploration_disable.png",
			wantClaimActive:  false,
			wantExploreLevel: 308,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldState := &domain.State{
				Accounts: []domain.Account{
					{
						Characters: []domain.Gamer{
							{},
						},
					},
				},
			}

			screen := "exploration"
			newState, err := an.AnalyzeAndUpdateState(tt.screenshot, oldState, rules[screen])
			assert.NoError(t, err)

			char := newState.Accounts[0].Characters[0]

			t.Logf("exploration.state.isClaimActive: %v", char.Exploration.State.IsClaimActive)
			t.Logf("exploration.level: %d", char.Exploration.Level)

			assert.Equal(t, tt.wantClaimActive, char.Exploration.State.IsClaimActive)
			assert.Equal(t, tt.wantExploreLevel, char.Exploration.Level)
		})
	}
}
