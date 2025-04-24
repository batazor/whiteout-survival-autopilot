package device

import (
	"context"

	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

func (d *Device) SwitchTo(ctx context.Context, profileIdx, gamerIdx int) error {
	// сбросить FSM в начальное состояние
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.AreaLookup, d.triggerEvaluator, d.ActiveGamer())

	if gamerIdx == 0 {
		d.NextProfile(profileIdx, gamerIdx)
	} else {
		d.NextGamer(profileIdx, gamerIdx)
	}
	return nil
}
