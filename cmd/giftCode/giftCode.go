package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

/*───────── файлы ─────────*/

const (
	devicesRel = "db/devices.yaml"
	codesRel   = "db/giftCodes.yaml"
	scriptRel  = "cmd/giftCode" // dir с redeem_code.py
)

/*───────── структура giftCodes.yaml ─────────*/

type giftCodes struct {
	Codes []struct {
		Name    string            `yaml:"name"`
		UserFor map[string]string `yaml:"userFor"` // uid -> status
	} `yaml:"codes"`
}

/*───────────────── main ─────────────────*/

func main() {
	cwd, _ := os.Getwd()
	scriptDir := filepath.Join(cwd, scriptRel)

	players := loadPlayers(filepath.Join(cwd, devicesRel))
	codes := loadCodes(filepath.Join(cwd, codesRel))

	for ci := range codes.Codes {
		code := &codes.Codes[ci]

		if code.UserFor == nil {
			code.UserFor = make(map[string]string)
		}

		fmt.Printf("\n=== Code: %s ===\n", code.Name)
		stop := false

		for _, g := range players.AllGamers() {
			uid := int64(g.ID)
			uidStr := strconv.FormatInt(uid, 10)

			if _, done := code.UserFor[uidStr]; done {
				continue
			}

			status := runPython(scriptDir, uid, code.Name)

			switch status {
			case "SUCCESS":
				fmt.Printf("✅ %s (%d) SUCCESS\n", g.Nickname, uid)
				code.UserFor[uidStr] = "SUCCESS"
			case "ALREADY_RECEIVED":
				fmt.Printf("ℹ️  %s (%d) ALREADY_RECEIVED\n", g.Nickname, uid)
				code.UserFor[uidStr] = "ALREADY_RECEIVED"
			case "CDK_NOT_FOUND":
				fmt.Printf("🚫 Код %s не существует – прекращаю обработку этого кода\n", code.Name)
				stop = true
			default:
				fmt.Printf("❌ %s (%d) ERROR\n", g.Nickname, uid)
				code.UserFor[uidStr] = "ERROR"
			}

			if stop {
				break
			}
			time.Sleep(time.Second)
		}
	}

	saveCodes(filepath.Join(cwd, codesRel), codes)
	fmt.Println("\n💾 giftCodes.yaml сохранён")
}

/*──────── helpers ────────*/

// cd cmd/giftCode && uv run redeem_code.py -c CODE --fid UID
func runPython(dir string, uid int64, code string) string {
	cmd := exec.Command("uv", "run", "redeem_code.py",
		"-c", code,
		"--fid", fmt.Sprint(uid),
	)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out

	if err := cmd.Run(); err != nil {
		return "ERROR"
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

func loadCodes(path string) giftCodes {
	b, err := os.ReadFile(path)
	chk(err)
	var gc giftCodes
	chk(yaml.Unmarshal(b, &gc))
	return gc
}

func saveCodes(path string, gc giftCodes) {
	b, err := yaml.Marshal(&gc)
	chk(err)
	chk(os.WriteFile(path, b, 0644))
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
