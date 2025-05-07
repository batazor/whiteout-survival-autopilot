package fsm

import (
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var mainMenuTransitionPaths = map[string]map[string][]TransitionStep{
	state.StateMainMenuCity: {
		state.StateMainCity: {
			{Action: "from_main_menu_city_to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuWilderness: {
			{Action: "to_main_menu_wilderness", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuBuilding1: {
			{Action: "to_main_menu_building_1", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuBuilding2: {
			{Action: "to_main_menu_building_2", Wait: 300 * time.Millisecond},
		},
		state.StateInfantryCityView: {
			{
				Action:  "to_main_menu_infantry",
				Wait:    300 * time.Millisecond,
				Trigger: "troops.infantry.state.isAvailable",
			},
		},
		state.StateLancerCityView: {
			{
				Action:  "to_main_menu_lancer",
				Wait:    300 * time.Millisecond,
				Trigger: "troops.lancer.state.isAvailable",
			},
		},
		state.StateMarksmanCityView: {
			{
				Action:  "to_main_menu_marksman",
				Wait:    300 * time.Millisecond,
				Trigger: "troops.marksman.state.isAvailable",
			},
		},
		state.StateMainMenuTechResearch: {
			{Action: "to_main_menu_tech_research", Wait: 300 * time.Millisecond},
		},
		state.StateVIP: {
			{Action: "to_vip", Wait: 300 * time.Millisecond},
		},
		state.StateExploration: {
			{Action: "to_exploration", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Action: "to_alliance_manage", Wait: 300 * time.Millisecond},
		},
	},
}
