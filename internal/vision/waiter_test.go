package vision_test

import (
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

// mockADB реализует adb.DeviceController для тестов
type mockADB struct {
	img image.Image
}

func (m *mockADB) ClickOCRResult(result *domain.OCRResult) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockADB) ListDevices() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockADB) SetActiveDevice(serial string) {
	//TODO implement me
	panic("implement me")
}

func (m *mockADB) GetActiveDevice() string {
	//TODO implement me
	panic("implement me")
}

func (m *mockADB) ClickRegion(name string, area *config.AreaLookup) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockADB) Swipe(x1 int, y1 int, x2 int, y2 int, durationMs time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockADB) Screenshot(path string) (image.Image, error) {
	// Сохраняем скриншот, чтобы эмулировать поведение
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_ = png.Encode(f, m.img)

	return m.img, nil
}

func TestWaitForText(t *testing.T) {
	// Загружаем фиксированное изображение
	f, err := os.Open("references/welcome_back.png")
	assert.NoError(t, err)
	defer f.Close()

	img, err := png.Decode(f)
	assert.NoError(t, err)

	// Создаем mock ADB
	mock := &mockADB{img: img}

	// Устанавливаем таймаут
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пытаемся найти слово "welcome"
	result, err := vision.WaitForText(ctx, mock, []string{"welcome"}, 500*time.Millisecond, image.Rectangle{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	t.Logf("Найдено: %s (%.2f)", result.Text, result.Confidence)
}
