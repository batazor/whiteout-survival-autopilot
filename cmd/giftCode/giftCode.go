// cmd/redeemer/main.go
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
	scriptRel  = "cmd/giftCode" // redeem_code.py lives here
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
	codesPath := filepath.Join(cwd, codesRel)

	players := loadPlayers(filepath.Join(cwd, devicesRel))
	codes := loadCodes(codesPath)

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

			// пропускаем, если прошлый статус не ERROR
			if prev, ok := code.UserFor[uidStr]; ok &&
				!strings.HasPrefix(prev, "ERROR REDEEM") &&
				!strings.HasPrefix(prev, "ERROR CAPTCHA_REQUEST") {
				continue
			}

			status := runPython(scriptDir, uid, code.Name)

			switch {
			case status == "SUCCESS":
				fmt.Printf("✅ %s (%d) SUCCESS\n", g.Nickname, uid)
			case status == "ALREADY_RECEIVED":
				fmt.Printf("ℹ️  %s (%d) ALREADY_RECEIVED\n", g.Nickname, uid)
			case status == "CDK_NOT_FOUND":
				fmt.Printf("🚫 Код %s не существует – прекращаю обработку этого кода\n", code.Name)
				stop = true
			default: // любой ERROR
				fmt.Printf("❌ %s (%d) %s\n", g.Nickname, uid, status)
			}

			// обновляем YAML в любом случае
			code.UserFor[uidStr] = status
			saveCodes(codesPath, codes)

			if stop {
				break
			}
			time.Sleep(time.Second)
		}
	}

	// дополнительная «страховка»
	saveCodes(codesPath, codes)
	fmt.Println("\n💾 giftCodes.yaml сохранён")
}

/*──────── helpers ────────*/

func runPython(dir string, uid int64, code string) string {
	cmd := exec.Command("uv", "run", "redeem_code.py",
		"-c", code,
		"--fid", fmt.Sprint(uid),
	)
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
