package fsm_test

import (
	"testing"

	"github.com/batazor/whiteout-survival-autopilot/src/fsm"
)

func TestFSMTransitions(t *testing.T) {
	tests := []struct {
		name     string
		events   []fsm.Event
		expected string
	}{
		{
			name: "Main city to AllianceManage and back to main_city",
			events: []fsm.Event{
				fsm.EventGoToAllianceManage,
				fsm.EventBack,
			},
			expected: fsm.StateMainCity,
		},
		{
			name: "Main city to Events to ActivityTriumph then back to Events",
			events: []fsm.Event{
				fsm.EventGoToEvents,
				fsm.EventGoToActivityTriumph,
				fsm.EventBack,
			},
			expected: fsm.StateEvents,
		},
		{
			name: "Main city to Profile and back to main_city",
			events: []fsm.Event{
				fsm.EventGoToProfile,
				fsm.EventBack,
			},
			expected: fsm.StateMainCity,
		},
		{
			name: "Main city to DawnMarket and back to main_city",
			events: []fsm.Event{
				fsm.EventGoToDawnMarket,
				fsm.EventBack,
			},
			expected: fsm.StateMainCity,
		},
		{
			name: "Main city to Exploration and back to main_city",
			events: []fsm.Event{
				fsm.EventGoToExploration,
				fsm.EventBack,
			},
			expected: fsm.StateMainCity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем новый экземпляр FSM с начальным состоянием main_city.
			gameFSM := fsm.NewGameFSM()

			// Обрабатываем последовательность событий.
			for _, event := range tc.events {
				err := gameFSM.Transition(event)
				if err != nil {
					t.Fatalf("Ошибка перехода по событию %s: %v", event, err)
				}
			}

			// Проверяем, соответствует ли конечное состояние ожидаемому.
			finalState := gameFSM.Current()
			if finalState != tc.expected {
				t.Errorf("Ожидалось состояние %s, получено %s", tc.expected, finalState)
			}
		})
	}
}
