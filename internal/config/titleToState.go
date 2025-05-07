package config

import (
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var (
	TitleToState = map[string][]string{
		"Alliance":      {state.StateAllianceManage},
		"Chief Profile": {state.StateChiefProfile},
		"Tech":          {state.StateAllianceTech},
		"Chests":        {state.StateAllianceChestGift, state.StateAllianceChestLoot},
		"Exploration":   {state.StateExploration},
		"Mail":          {state.StateMail, state.StateMailWars, state.StateMailAlliance, state.StateMailSystem, state.StateMailReports, state.StateMailStarred},
	}
)

func SameScreenGroup(current, other string) bool {
	for _, states := range TitleToState {
		foundCurrent := false
		foundOther := false
		for _, st := range states {
			if st == current {
				foundCurrent = true
			}
			if st == other {
				foundOther = true
			}
		}
		if foundCurrent && foundOther {
			return true
		}
	}
	return false
}
