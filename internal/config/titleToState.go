package config

import (
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

var (
	// TitleToState используется для определения состояния (state) в зависимости от названия экрана (title).
	TitleToState = map[string][]string{
		"Alliance":      {state.StateAllianceManage},
		"Chief Profile": {state.StateChiefProfile},
		"Tech":          {state.StateAllianceTech},
		"Chests":        {state.StateAllianceChests, state.StateAllianceChestGift, state.StateAllianceChestLoot},
		"Exploration":   {state.StateExploration},
		"Mail":          {state.StateMail, state.StateMailWars, state.StateMailAlliance, state.StateMailSystem, state.StateMailReports, state.StateMailStarred},
		"Backpack":      {state.StateBackpack, state.StateBackpackResources, state.StateBackpackSpeedups, state.StateBackpackBonus, state.StateBackpackGear, state.StateBackpackOther},
		"Chat":          {state.StateChat, state.StateChatAlliance, state.StateChatWorld, state.StateChatPersonal},
		"Heroes":        {state.StateHeroes},
		"Events":        {state.StateEvents},
		"Deals":         {state.StateDeals},
		"War":           {state.StateAllianceWar, state.StateAllianceWarAutoJoin},
		"VIP":           {state.StateVIP},
		"Settings":      {state.StateChiefProfileSetting},
		"Account":       {state.StateChiefProfileAccount},
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
