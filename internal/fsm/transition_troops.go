package fsm

import (
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var troopsTransitionPaths = map[string]map[string][]TransitionStep{
	state.StateInfantryCityView: {
		state.StateMainCity: {},
		state.StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateLancerCityView: {
		state.StateMainCity: {},
		state.StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMarksmanCityView: {
		state.StateMainCity: {},
		state.StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
}
