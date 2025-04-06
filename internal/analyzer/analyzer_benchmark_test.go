package analyzer

import (
	"log/slog"
	"os"
	"testing"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func BenchmarkAnalyzeAndUpdateState(b *testing.B) {
	// --- Prepare analyzer ---
	areaConfig, err := config.LoadAreaReferences("../../references/area.json")
	if err != nil {
		b.Fatalf("failed to load area.json: %v", err)
	}

	rules, err := config.LoadAnalyzeRules("../../references/analyze.yaml")
	if err != nil {
		b.Fatalf("failed to load analyze.yaml: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	an := NewAnalyzer(areaConfig, logger)

	state := &domain.Gamer{ID: 123}

	imagePath := "../../references/screenshots/city_main.png"
	screen := "main_city"

	// --- Reset timer and benchmark ---
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := an.AnalyzeAndUpdateState(imagePath, state, rules[screen])
		if err != nil {
			b.Errorf("analysis failed: %v", err)
		}
	}
}
