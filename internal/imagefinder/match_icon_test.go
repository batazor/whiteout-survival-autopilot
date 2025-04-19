package imagefinder

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gocv.io/x/gocv"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

type iconCase struct {
	name          string
	screenshot    string
	icon          string
	regionKey     string
	threshold     float32
	wantMatch     bool
	minConfidence float32
}

func TestMatchIconInRegion(t *testing.T) {
	// ─── Загрузка area.json один раз ────────────────────────────────────────────
	areaConfig, err := config.LoadAreaReferences("../../references/area.json")
	assert.NoError(t, err)

	// ─── Логгер в файл ──────────────────────────────────────────────────────────
	logPath := "../../logs/test_match_icon_table.log"
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	logFile, err := os.Create(logPath)
	assert.NoError(t, err)
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	// ─── Набор тест‑кейсов ──────────────────────────────────────────────────────
	tests := []iconCase{
		{
			name:          "NeedSupport – есть иконка",
			screenshot:    "../../references/screenshots/city_main.png",
			icon:          "../../references/icons/alliance.state.isNeedSupport.png",
			regionKey:     "alliance.state.isNeedSupport",
			threshold:     0.4,
			wantMatch:     true,
			minConfidence: 0.4,
		},
		{
			name:          "NoNeedSupport – иконки нет",
			screenshot:    "../../references/screenshots/alliance.png",
			icon:          "../../references/icons/alliance.state.isNeedSupport.png",
			regionKey:     "alliance.state.isNeedSupport",
			threshold:     0.4,
			wantMatch:     false,
			minConfidence: 0.0, // можно не проверять confidence
		},
		{
			name:          "ClaimButton – есть зелёная кнопка",
			screenshot:    "../../references/screenshots/alliance_chest_gift.png",
			icon:          "../../references/icons/alliance.state.isClaimButton.png",
			regionKey:     "alliance.state.isClaimButton",
			threshold:     0.3,
			wantMatch:     true,
			minConfidence: 0.1,
		},
		// добавляйте новые случаи сюда …
	}

	// ─── Запуск ────────────────────────────────────────────────────────────────
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, ok := areaConfig.Get(tt.regionKey)
			assert.Truef(t, ok, "region %s not found in area.json", tt.regionKey)

			match, conf, err := MatchIconInRegion(
				tt.screenshot,
				tt.icon,
				region.Zone,
				tt.threshold,
				logger,
			)
			assert.NoError(t, err)

			if tt.minConfidence > 0 {
				t.Logf("[%s] confidence = %.3f", tt.name, conf)
				assert.GreaterOrEqual(t, conf, tt.minConfidence)
			}
			assert.Equal(t, tt.wantMatch, match)

			// ─── Отладочная визуализация найденных совпадений ─────────────
			debugPath := filepath.Join("../../out", "debug_"+tt.name+".png")
			err = saveMatchDebugImage(tt.screenshot, tt.icon, region.Zone, tt.threshold, debugPath)
			assert.NoError(t, err, "failed to save debug match image")
			t.Logf("Saved debug image: %s", debugPath)
		})
	}
}

// saveMatchDebugImage выполняет MatchTemplate и сохраняет PNG с подсветкой:
// - зелёным — если score ≥ threshold
// - красным  — если 0 < score < threshold
func saveMatchDebugImage(screenshotPath, iconPath string, region image.Rectangle, threshold float32, debugPath string) error {
	screenshotMat := gocv.IMRead(screenshotPath, gocv.IMReadColor)
	if screenshotMat.Empty() {
		return fmt.Errorf("cannot read screenshot: %s", screenshotPath)
	}
	defer screenshotMat.Close()

	iconMat := gocv.IMRead(iconPath, gocv.IMReadColor)
	if iconMat.Empty() {
		return fmt.Errorf("cannot read icon: %s", iconPath)
	}
	defer iconMat.Close()

	if region.Empty() {
		region = image.Rect(0, 0, screenshotMat.Cols(), screenshotMat.Rows())
	}
	crop := screenshotMat.Region(region)
	defer crop.Close()

	result := gocv.NewMat()
	defer result.Close()

	gocv.MatchTemplate(crop, iconMat, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	overlay := crop.Clone()
	defer overlay.Close()

	rows, cols := result.Rows(), result.Cols()
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			score := result.GetFloatAt(y, x)
			if score <= 0 {
				continue
			}

			rect := image.Rect(x, y, x+iconMat.Cols(), y+iconMat.Rows())
			var col color.RGBA
			if score >= threshold {
				col = color.RGBA{G: 255, A: 255} // зелёный
			} else {
				col = color.RGBA{R: 255, A: 255} // красный
			}
			gocv.Rectangle(&overlay, rect, col, 2)
		}
	}

	if ok := gocv.IMWrite(debugPath, overlay); !ok {
		return fmt.Errorf("failed to write debug image to: %s", debugPath)
	}
	return nil
}
