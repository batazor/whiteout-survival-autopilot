package fsm_test

import (
	"image"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

// FakeADB реализует интерфейс adb.DeviceController и записывает вызовы ClickRegion.
type FakeADB struct {
	Clicks []string
}

func (f *FakeADB) Screenshot(path string) (image.Image, error) {
	file, err := os.Open("../../references/screenshots/main_city/alliance_need_support.png")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	return img, err
}
func (f *FakeADB) ClickOCRResult(result *domain.OCRResult) error {
	panic("not implemented")
}
func (f *FakeADB) Swipe(x1 int, y1 int, x2 int, y2 int, durationMs time.Duration) error {
	panic("not implemented")
}
func (f *FakeADB) ClickRegion(name string, lookup *config.AreaLookup) error {
	f.Clicks = append(f.Clicks, name)
	return nil
}
func (f *FakeADB) ListDevices() ([]string, error) { return []string{"fake"}, nil }
func (f *FakeADB) SetActiveDevice(serial string)  {}
func (f *FakeADB) GetActiveDevice() string        { return "fake" }
func (f *FakeADB) RestartApplication() error {
	return nil
}
func (f *FakeADB) Click(region image.Rectangle) error {
	return nil
}

func TestForceTo(t *testing.T) {
	lookup, err := config.LoadAreaReferences("../../references/area.json")
	if err != nil {
		t.Fatalf("failed to load area.json: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	fakeADB := &FakeADB{}

	tests := []struct {
		name           string
		target         string
		expectedClicks []string
	}{
		{
			name:           "Switch to ChiefCharacters",
			target:         state.StateChiefCharacters,
			expectedClicks: []string{"to_chief_profile", "to_chief_profile_setting", "to_chief_characters"},
		},
		{
			name:           "Switch to ChiefProfileAccountChangeGoogle",
			target:         state.StateChiefProfileAccountChangeGoogle,
			expectedClicks: []string{"to_chief_profile", "to_chief_profile_setting", "to_chief_profile_account", "to_change_account", "to_google_account"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем FSM; начальное состояние – StateMainCity.
			gameFSM := fsm.NewGame(logger, fakeADB, lookup, nil, nil)

			// Сбросим накопленные клики перед каждым тестом.
			fakeADB.Clicks = nil
			gameFSM.ForceTo(tc.target, nil)

			if len(fakeADB.Clicks) != len(tc.expectedClicks) {
				t.Errorf("expected %d clicks, got %d: %v", len(tc.expectedClicks), len(fakeADB.Clicks), fakeADB.Clicks)
			}
			for i, expected := range tc.expectedClicks {
				if fakeADB.Clicks[i] != expected {
					t.Errorf("expected click %q at position %d, got %q", expected, i, fakeADB.Clicks[i])
				}
			}
		})
	}
}
