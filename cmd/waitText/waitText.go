package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
)

func main() {
	// 1) Инициализируем клиент
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	client := ocrclient.NewClient("RF8RC00M8MF", logger)

	// 2) Задаём слова-стопы, таймаут и интервал опроса
	stopWords := []string{"Chief Profile"}
	timeout := 30 * time.Second // максимум ждать 30 секунд
	interval := 1 * time.Second // проверять экран каждую 1 секунду
	debugName := "screenState.titleFact"

	// 3) Запускаем ожидание
	results, err := client.WaitForText(stopWords, timeout, interval, debugName)
	if err != nil {
		logger.Error("WaitForText failed", "error", err)
		return
	}

	// 4) Обрабатываем результат — список OCR-зон, где найден текст
	if len(results) == 0 {
		fmt.Println("⚠️ Ни одно из стоп-слов не найдено")
		return
	}

	fmt.Println("✅ Нашли одно из слов! Зоны с распознанным текстом:")
	for i, res := range results {
		fmt.Printf("  %d) \"%s\" (%.2f) @ X:%d Y:%d W:%d H:%d\n",
			i+1, res.Text, res.Score, res.X, res.Y, res.Width, res.Height)
	}
}
