package fsm

import (
	"time"
)

const (
	// Tundra Adventure
	StateTundraAdventure               = "tundra_adventure"
	StateTundraAdventureMain           = "tundra_adventure_main"
	StateTundraAdventureDrill          = "tundra_adventure_drill"
	StateTundraAdventurerDrill         = "tundra_adventurer_drill"
	StateTundraAdventurerDailyMissions = "tundra_adventurer_daily_missions"
	StateTundraAdventureOdessey        = "tundra_adventure_odessey"
	StateTundraAdventureCaravan        = "tundra_adventure_caravan"
)

var tundraAdventureTransitionPaths = map[string]map[string][]TransitionStep{
	StateTundraAdventure: {
		StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_main", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureDrill: {
			{Action: "to_tundra_adventure_drill", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
		StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	StateTundraAdventureMain: {
		StateTundraAdventureDrill: {
			{Action: "to_tundra_adventure_drill", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
		StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	StateTundraAdventureDrill: {
		StateTundraAdventurerDrill: {
			{Action: "to_tundra_adventurer_drill", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventurerDailyMissions: {
			{Action: "to_tundra_adventurer_daily_missions", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	StateTundraAdventurerDrill: {
		StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventurerDailyMissions: {
			{Action: "to_tundra_adventurer_daily_missions", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureOdessey: {
			{Action: "to_tundra_adventure_odessey", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventureCaravan: {
			{Action: "to_tundra_adventure_caravan", Wait: 300 * time.Millisecond},
		},
	},
	StateTundraAdventureOdessey: {
		StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
	StateTundraAdventureCaravan: {
		StateTundraAdventureMain: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
		StateMainCity: {
			{Action: "to_tundra_adventure_back", Wait: 300 * time.Millisecond},
		},
	},
}
