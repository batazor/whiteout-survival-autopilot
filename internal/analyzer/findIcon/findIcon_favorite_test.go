package finder

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindIcons(t *testing.T) {
	// Задаем пути к изображениям
	screenshotPath := "testdata/alliance_tech_2.png"
	iconPath := "testdata/icon.png"

	// Логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Запускаем функцию поиска
	threshold := float32(0.9)
	boxes, err := FindIcons(screenshotPath, iconPath, threshold, logger)
	if err != nil {
		t.Fatalf("ошибка при выполнении FindIcons: %v", err)
	}

	// Проверяем, что хотя бы одна иконка найдена
	if len(boxes) == 0 {
		t.Error("иконки не найдены, хотя ожидалось хотя бы одно совпадение")
	}

	err = DrawBoxes(screenshotPath, boxes)
	require.Nil(t, err, "ошибка при отрисовке иконок на изображении")

	t.Logf("Найдено иконок: %d", len(boxes))
}
