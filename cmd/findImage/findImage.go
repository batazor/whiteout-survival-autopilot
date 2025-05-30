package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
)

func main() {
	// 1) Создаём OCR-клиент
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	client := ocrclient.NewClient("RF8RC00M8MF", logger)

	// 2) Пытаемся найти иконку "alliance.state.isNeedSupport" (имя файла alliance.state.isNeedSupport.png в references/icons),
	//    c порогом 0.8 и меткой debug_name "alliance.state.isNeedSupport_check"
	start := time.Now()
	resp, err := client.FindImage("alliance.state.isNeedSupport", 0.8, "alliance.state.isNeedSupport_check")
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("FindImage failed: %v", err)
	}

	// 3) Разбираем результат
	if resp.Found {
		// конвертируем в прямоугольники image.Rectangle
		rects := resp.ToRects()
		fmt.Printf("✅ Найдена иконка «alliance.state.isNeedSupport» (threshold=0.8) за %v:\n", elapsed)
		for i, r := range rects {
			fmt.Printf("  #%d at x=%d,y=%d – x2=%d,y2=%d\n",
				i+1, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
		}
	} else {
		fmt.Printf("❌ Иконка «alliance.state.isNeedSupport» не найдена (checked за %v)\n", elapsed)
	}
}
