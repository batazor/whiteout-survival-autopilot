package finder

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestFindIcons(t *testing.T) {
	// Задаем пути к изображениям
	screenshotPath := filepath.Join("testdata", "alliance_tech_2.png")
	iconPath := filepath.Join("testdata", "icon.png")

	// Логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Запускаем функцию поиска
	threshold := float32(0.5)
	boxes, err := FindIcons(screenshotPath, iconPath, threshold, logger)
	if err != nil {
		t.Fatalf("ошибка при выполнении FindIcons: %v", err)
	}

	// Проверяем, что хотя бы одна иконка найдена
	if len(boxes) == 0 {
		t.Error("иконки не найдены, хотя ожидалось хотя бы одно совпадение")
	}

	t.Logf("Найдено иконок: %d", len(boxes))
}
