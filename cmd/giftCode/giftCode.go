// cmd/redeemer/main.go
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

/*───────────── файлы ─────────────*/

const (
	devicesRel = "db/devices.yaml"   // список игроков
	codesRel   = "db/giftCodes.yaml" // коды + userFor
	scriptRel  = "cmd/giftCode"      // <─ новый путь к redeem_code.py
)

/*───────────── структура codes.yaml ─────────────*/

type giftCodes struct {
	Codes []struct {
		Name    string  `yaml:"name"`
		UserFor []int64 `yaml:"userFor"`
	} `yaml:"codes"`
}

/*───────────────── main ─────────────────*/

func main() {
	cwd, _ := os.Getwd()
	scriptDir := filepath.Join(cwd, scriptRel)

	players := loadPlayers(filepath.Join(cwd, devicesRel))
	codes := loadCodes(filepath.Join(cwd, codesRel))

	for ci, code := range codes.Codes {
		claimed := make(map[int64]struct{}, len(code.UserFor))
		for _, id := range code.UserFor {
			claimed[id] = struct{}{}
		}

		fmt.Printf("\n=== Code: %s ===\n", code.Name)
		stop := false

		for _, g := range players.AllGamers() {
			uid := int64(g.ID)
			if _, done := claimed[uid]; done {
				continue
			}

			status := runPython(scriptDir, uid, code.Name)

			switch status {
			case "SUCCESS":
				fmt.Printf("✅ %s (%d) SUCCESS\n", g.Nickname, uid)
				codes.Codes[ci].UserFor = append(codes.Codes[ci].UserFor, uid)
			case "ALREADY_RECEIVED":
				fmt.Printf("ℹ️  %s (%d) ALREADY_RECEIVED\n", g.Nickname, uid)
				codes.Codes[ci].UserFor = append(codes.Codes[ci].UserFor, uid)
			case "CDK_NOT_FOUND":
				fmt.Printf("🚫 Код %s не существует – прекращаю обработку этого кода\n", code.Name)
				stop = true
			default:
				fmt.Printf("❌ %s (%d) ERROR\n", g.Nickname, uid)
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

/*───────────── helpers ─────────────*/

// runPython → cd cmd/giftCode && uv run redeem_code.py -c CODE --fid UID
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
