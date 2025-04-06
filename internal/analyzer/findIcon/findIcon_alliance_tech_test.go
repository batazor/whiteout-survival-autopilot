package finder

import (
	"log/slog"
	"os"
	"testing"
)

func TestFindAllianceTechIcons(t *testing.T) {
	// Логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Запускаем функцию поиска
	threshold := float32(0.4)
	screen := "../../../references/screenshots/alliance.png"
	icon := "../../../references/icons/alliance.state.isAllianceTechButton.png"

	boxes, err := FindIcons(screen, icon, threshold, logger)
	if err != nil {
		t.Fatalf("ошибка при выполнении FindIcons: %v", err)
	}

	// Проверяем, что хотя бы одна иконка найдена
	if len(boxes) == 0 {
		t.Error("иконки не найдены, хотя ожидалось хотя бы одно совпадение")
	}

	t.Logf("Найдено иконок: %d", len(boxes))
}
