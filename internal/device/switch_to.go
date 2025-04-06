package device

import (
	"context"
)

func (d *Device) SwitchTo(ctx context.Context, profileIdx, gamerIdx int) error {
	if gamerIdx == 0 {
		d.NextProfile(profileIdx, gamerIdx)
	} else {
		d.NextGamer(profileIdx, gamerIdx)
	}
	return nil
}
