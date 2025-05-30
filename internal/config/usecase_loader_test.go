package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

func writeUseCase(t *testing.T, dir, filename, content string) {
	t.Helper()
	full := filepath.Join(dir, filename)
	err := os.MkdirAll(filepath.Dir(full), 0755)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(full, []byte(content), 0644))
}

func TestLoadAll_CronAndDebug(t *testing.T) {
	tmpDir := t.TempDir()

	writeUseCase(t, tmpDir, "cron.yaml", `
name: FromCron
cron: "* * * * *"
priority: 1
node: main_city
steps: []
`)

	writeUseCase(t, tmpDir, "debug/only_debug.yaml", `
name: FromDebug
priority: 1
node: main_city
steps: []
`)

	writeUseCase(t, tmpDir, "skipped.yaml", `
name: Skipped
priority: 1
node: main_city
steps: []
`)

	loader := config.NewUseCaseLoader(tmpDir)
	active, err := loader.LoadAll(context.Background())
	require.NoError(t, err)

	// Активные: только те, что с cron или из debug
	require.Len(t, active, 2)
	names := map[string]bool{}
	for _, uc := range active {
		names[uc.Name] = true
	}
	require.True(t, names["FromCron"])
	require.True(t, names["FromDebug"])
	require.False(t, names["Skipped"])

	// В индексе — все три
	require.NotNil(t, loader.GetByName("FromCron"))
	require.NotNil(t, loader.GetByName("FromDebug"))
	require.NotNil(t, loader.GetByName("Skipped"))
	require.Nil(t, loader.GetByName("Unknown"))

	// TTL проставлен для debug usecase
	fromDebug := loader.GetByName("FromDebug")
	require.Equal(t, time.Duration(1), fromDebug.TTL)

	// SourcePath должен быть сохранён
	require.Contains(t, fromDebug.SourcePath, "debug/only_debug.yaml")
}
