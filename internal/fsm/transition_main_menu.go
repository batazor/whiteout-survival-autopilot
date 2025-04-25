package fsm

import (
	"time"
)

const (
	// Главное меню
	StateMainMenuCity         = "main_menu_city"
	StateMainMenuWilderness   = "main_menu_wilderness"
	StateMainMenuBuilding1    = "main_menu_building_1"
	StateMainMenuBuilding2    = "main_menu_building_2"
	StateMainMenuTechResearch = "main_menu_tech_research"
)

var mainMenuTransitionPaths = map[string]map[string][]TransitionStep{
	StateMainMenuCity: {
		StateMainMenuWilderness: {
			{Action: "to_main_menu_wilderness", Wait: 300 * time.Millisecond},
		},
		StateMainMenuBuilding1: {
			{Action: "to_main_menu_building_1", Wait: 300 * time.Millisecond},
		},
		StateMainMenuBuilding2: {
			{Action: "to_main_menu_building_2", Wait: 300 * time.Millisecond},
		},
		StateInfantryCityView: {
			{
				Action:  "to_main_menu_infantry",
				Wait:    300 * time.Millisecond,
				Trigger: "troops.infantry.state.isAvailable",
			},
		},
		StateLancerCityView: {
			{
				Action:  "to_main_menu_lancer",
				Wait:    300 * time.Millisecond,
				Trigger: "troops.lancer.state.isAvailable",
			},
		},
		StateMarksmanCityView: {
			{
				Action:  "to_main_menu_marksman",
				Wait:    300 * time.Millisecond,
				Trigger: "troops.marksman.state.isAvailable",
			},
		},
		StateMainMenuTechResearch: {
			{Action: "to_main_menu_tech_research", Wait: 300 * time.Millisecond},
		},
	},
}
