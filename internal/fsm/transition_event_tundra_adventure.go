package fsm

import (
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var tundraAdventureTransitionPaths = map[string]map[string][]TransitionStep{
	state.StateTundraAdventure: {
		state.StateTundraAdventureMain: {
			{Click: "to_tundra_adventure_main", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureDrill: {
			{Click: "to_tundra_adventure_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Click: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Click: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureMain: {
		state.StateTundraAdventureDrill: {
			{Click: "to_tundra_adventure_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Click: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Click: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureDrill: {
		state.StateTundraAdventurerDrill: {
			{Click: "to_tundra_adventurer_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventurerDailyMissions: {
			{Click: "to_tundra_adventurer_daily_missions", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureMain: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventurerDrill: {
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventurerDailyMissions: {
			{Click: "to_tundra_adventurer_daily_missions", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureMain: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Click: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Click: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventurerDailyMissions: {
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventurerDrill: {
			{Click: "to_tundra_adventurer_drill", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureMain: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureOdessey: {
			{Click: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventureCaravan: {
			{Click: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureOdessey: {
		state.StateTundraAdventureMain: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateTundraAdventureCaravan: {
		state.StateTundraAdventureMain: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
}
