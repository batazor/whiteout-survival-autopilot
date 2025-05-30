package main

import (
	"path/filepath"

	"github.com/batazor/whiteout-survival-autopilot/internal/gift"
)

func main() {
	gift.RunRedeemer(gift.RedeemConfig{
		DevicesYAML: filepath.Join("db", "devices.yaml"),
		CodesYAML:   filepath.Join("db", "giftCodes.yaml"),
		// PythonDir: ""  // пусто => redeem_code.py из internal/discordgift
	})
}
