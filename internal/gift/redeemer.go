package gift

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

/*──────── public API ───────*/

type RedeemConfig struct {
	DevicesYAML string // db/devices.yaml
	CodesYAML   string // db/giftCodes.yaml
	PythonDir   string // каталог со скриптом redeem_code.py ("" ⇒ пакет discordgift)
}

func RunRedeemer(cfg RedeemConfig) {
	// ── путь к redeem_code.py ──
	if cfg.PythonDir == "" {
		_, thisFile, _, _ := runtime.Caller(0) // …/internal/discordgift/redeemer.go
		cfg.PythonDir = filepath.Dir(thisFile) // …/internal/discordgift
	}

	players := loadPlayers(cfg.DevicesYAML)
	codes := loadCodes(cfg.CodesYAML)

	for ci := range codes.Codes {
		code := &codes.Codes[ci]
		if code.UserFor == nil {
			code.UserFor = map[string]string{}
		}

		fmt.Printf("\n=== Code: %s ===\n", code.Name)
		stop := false

		for _, g := range players.AllGamers() {
			uidStr := strconv.FormatInt(int64(g.ID), 10)

			// пропускаем, если не ошибка
			if prev, ok := code.UserFor[uidStr]; ok &&
				!strings.HasPrefix(prev, "ERROR REDEEM") &&
				!strings.HasPrefix(prev, "ERROR CAPTCHA_REQUEST") {
				continue
			}

			status := runPython(cfg.PythonDir, int64(g.ID), code.Name)

			switch {
			case status == "SUCCESS":
				fmt.Printf("✅ %s (%s) SUCCESS\n", g.Nickname, uidStr)
			case status == "ALREADY_RECEIVED":
				fmt.Printf("ℹ️  %s (%s) ALREADY_RECEIVED\n", g.Nickname, uidStr)
			case status == "CDK_NOT_FOUND":
				fmt.Printf("🚫 Код %s не существует – стоп\n", code.Name)
				stop = true
			default:
				fmt.Printf("❌ %s (%s) %s\n", g.Nickname, uidStr, status)
			}

			code.UserFor[uidStr] = status   // записываем всегда
			saveCodes(cfg.CodesYAML, codes) // и сразу сохраняем

			if stop {
				break
			}
			time.Sleep(time.Second)
		}
	}
	fmt.Println("\n💾 giftCodes.yaml сохранён")
}

/*──────── helpers ───────*/

func runPython(dir string, uid int64, code string) string {
	cmd := exec.Command("uv", "run", "redeem_code.py",
		"-c", code, "--fid", fmt.Sprint(uid))
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		return "ERROR UV_RUN"
	}
	return strings.TrimSpace(out.String())
}

func loadPlayers(path string) *domain.Config {
	b, err := os.ReadFile(path)
	chk(err)
	var cfg domain.Config
	chk(yaml.Unmarshal(b, &cfg))
	return &cfg
}

func loadCodes(path string) domain.GiftCodes {
	b, err := os.ReadFile(path)
	chk(err)
	var gc domain.GiftCodes
	chk(yaml.Unmarshal(b, &gc))
	return gc
}

func saveCodes(path string, gc domain.GiftCodes) {
	b, err := yaml.Marshal(&gc)
	chk(err)
	chk(os.WriteFile(path, b, 0o644))
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
