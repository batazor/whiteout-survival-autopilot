package fsm_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	fsm2 "github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

func TestExpectState_TableDriven(t *testing.T) {
	// Set ENV
	os.Setenv("PATH_TO_FSM_STATE_RULES", "../../references/fsmState.yaml")

	lookup, err := config.LoadAreaReferences("../../references/area.json")
	if err != nil {
		t.Fatalf("failed to load area.json: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	fakeADB := &FakeADB{}

	tests := []struct {
		name           string
		want           string
		ocrTitle       string
		ocrFamily      string
		expectedResult string
		expectedErr    bool
		analyzeErr     error
	}{
		{
			name:           "want входит в фильтрованную группу (город)",
			want:           "main_city",
			ocrTitle:       "MainCity", // <-- ключ совпадает с мапой!
			ocrFamily:      "world",    // фильтруем на main_city
			expectedResult: "main_city",
		},
		{
			name:           "want входит в фильтрованную группу (мир)",
			want:           "world",
			ocrTitle:       "MainCity",
			ocrFamily:      "city", // фильтруем на world
			expectedResult: "world",
		},
		{
			name:           "want не в группе, но группа не пуста — вернёт первый",
			want:           "some_state",
			ocrTitle:       "MainCity",
			ocrFamily:      "",
			expectedResult: "some_state",
		},
		{
			name:           "нет совпадений по заголовку — возвращает want",
			want:           "mail",
			ocrTitle:       "Unknown", // Нет такого ключа
			ocrFamily:      "",
			expectedResult: "mail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTitleToState := config.TitleToState
			defer func() { config.TitleToState = oldTitleToState }()

			gamer := &domain.Gamer{
				ScreenState: domain.ScreenState{},
			}
			gameFSM := fsm2.NewGame(logger, fakeADB, lookup, nil, gamer)

			got, err := gameFSM.ExpectState(tt.want)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, got)
			}
		})
	}
}
