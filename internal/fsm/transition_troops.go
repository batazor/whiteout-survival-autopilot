package fsm

import (
	"time"
)

const (
	StateInfantryCityView = "infantry_city_view"
	StateLancerCityView   = "lancer_city_view"
	StateMarksmanCityView = "marksman_city_view"
)

var troopsTransitionPaths = map[string]map[string][]TransitionStep{
	StateInfantryCityView: {
		StateMainCity: {},
		StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	StateLancerCityView: {
		StateMainCity: {},
		StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	StateMarksmanCityView: {
		StateMainCity: {},
		StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
}
