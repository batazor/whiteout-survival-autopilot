package device_test

import (
	"context"
	"image"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

type MockADB struct {
	imagePath string
}

func (m *MockADB) Screenshot(path string) (image.Image, error) {
	return nil, nil
}

func (m *MockADB) ClickRegion(name string, lookup *config.AreaLookup) error {
	return nil
}

func (m *MockADB) ClickOCRResult(_ *domain.OCRResult) error {
	return nil
}

func (m *MockADB) Swipe(x1 int, y1 int, x2 int, y2 int, durationMs time.Duration) error {
	return nil
}

func (m *MockADB) ListDevices() ([]string, error) {
	return nil, nil
}

func (m *MockADB) SetActiveDevice(serial string) {}

func (m *MockADB) GetActiveDevice() string {
	return ""
}

func TestDetectedGamer_WithRealConfig_AndDeviceNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 🧾 Загружаем конфиг
	cfg, err := config.LoadDeviceConfig("../../db/devices.yaml", "../../db/state.yaml")
	if err != nil {
		t.Fatalf("❌ Не удалось загрузить devices.yaml: %v", err)
	}

	if len(cfg.Devices) == 0 || len(cfg.Devices[0].Profiles) == 0 {
		t.Fatal("❌ В конфиге нет устройств или профилей")
	}

	// ⚙️ Подменяем ADB-контроллер на мок
	profiles := cfg.Devices[0].Profiles
	log := slog.Default()

	// Создаём Device через `New`, потом подменяем ADB и FSM
	dev, err := device.New("test-device", profiles, log, "../../references/area.json")
	if err != nil {
		t.Fatalf("❌ device.New() вернул ошибку: %v", err)
	}

	// Load area.json
	lookup, err := config.LoadAreaReferences("../../references/area.json")
	if err != nil {
		t.Fatalf("failed to load area.json: %v", err)
	}

	dev.FSM = fsm.NewGame(log, &MockADB{imagePath: "../../references/screenshots/chief_profile.png"}, lookup)

	// Подменяем только Screenshot-логику, остальное можно оставить
	dev.ADB = &MockADB{imagePath: "../../references/screenshots/chief_profile.png"}

	// 🚀 Выполняем обнаружение
	profileIdx, gamerIdx, err := dev.DetectedGamer(ctx, "../../references/screenshots/chief_profile.png")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, profileIdx, 0, "должен найти profile")
	assert.GreaterOrEqual(t, gamerIdx, 0, "должен найти gamer")

	nickname := dev.Profiles[profileIdx].Gamer[gamerIdx].Nickname
	t.Logf("✅ Найден игрок: profileIdx=%d, gamerIdx=%d, nickname=%s", profileIdx, gamerIdx, nickname)

	assert.Equal(t, "batazor", nickname)
}
