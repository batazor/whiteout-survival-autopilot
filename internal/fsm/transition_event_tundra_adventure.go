package fsm

import (
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var tundraAdventureTransitionPaths = map[string]map[string][]TransitionStep{
	state.StateTundraAdventure: {
		state.StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_main", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureDrill: {
			{Action: "to_tundra_adventure_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureMain: {
		state.StateTundraAdventureDrill: {
			{Action: "to_tundra_adventure_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureDrill: {
		state.StateTundraAdventurerDrill: {
			{Action: "to_tundra_adventurer_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventurerDailyMissions: {
			{Action: "to_tundra_adventurer_daily_missions", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventurerDrill: {
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventurerDailyMissions: {
			{Action: "to_tundra_adventurer_daily_missions", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventurerDailyMissions: {
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventurerDrill: {
			{Action: "to_tundra_adventurer_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureOdessey: {
		state.StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureCaravan: {
		state.StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
}
