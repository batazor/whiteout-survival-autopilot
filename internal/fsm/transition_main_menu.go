package fsm

import (
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var mainMenuTransitionPaths = map[string]map[string][]TransitionStep{
	state.StateMainMenuCity: {
		state.StateMainCity: {
			{Click: "from_main_menu_city_to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuWilderness: {
			{Click: "to_main_menu_wilderness", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuBuilding1: {
			{Click: "to_main_menu_building_1", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuBuilding2: {
			{Click: "to_main_menu_building_2", Wait: 300 * time.Millisecond},
		},
		state.StateInfantryCityView: {
			{
				Click: "to_main_menu_infantry",
				Wait:  300 * time.Millisecond,
			},
		},
		state.StateLancerCityView: {
			{
				Click: "to_main_menu_lancer",
				Wait:  300 * time.Millisecond,
			},
		},
		state.StateMarksmanCityView: {
			{
				Click: "to_main_menu_marksman",
				Wait:  300 * time.Millisecond,
			},
		},
		state.StateMainMenuTechResearch: {
			{Click: "to_main_menu_tech_research", Wait: 300 * time.Millisecond},
		},
		state.StateVIP: {
			{Click: "to_vip", Wait: 300 * time.Millisecond},
		},
		state.StateExploration: {
			{Click: "to_exploration", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Click: "to_alliance_manage", Wait: 300 * time.Millisecond},
		},
		state.StateChiefProfile: {
			{Click: "to_chief_profile", Wait: 300 * time.Millisecond},
		},
	},
}
