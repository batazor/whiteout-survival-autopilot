package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
)

func main() {
	// 1) Загружаем все зоны из JSON
	lookup, err := config.LoadAreaReferences("references/area.json")
	if err != nil {
		log.Fatalf("failed to load area references: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 2) Берём первую зону "vip.state.isNotify"
	regionVIPIsNotifyRaw, ok := lookup.Get("vip.state.isNotify")
	if !ok {
		log.Fatal("area 'vip' not found")
	}

	regionVIPIsNotify := ocrclient.Region{
		X0: regionVIPIsNotifyRaw.Zone.Min.X,
		Y0: regionVIPIsNotifyRaw.Zone.Min.Y,
		X1: regionVIPIsNotifyRaw.Zone.Max.X,
		Y1: regionVIPIsNotifyRaw.Zone.Max.Y,
	}

	// 2.1) Берём вторую зону "gems"
	regionGemsRaw, ok := lookup.Get("gems")
	if !ok {
		log.Fatal("area 'gems' not found")
	}

	regionGems := ocrclient.Region{
		X0: regionGemsRaw.Zone.Min.X,
		Y0: regionGemsRaw.Zone.Min.Y,
		X1: regionGemsRaw.Zone.Max.X,
		Y1: regionGemsRaw.Zone.Max.Y,
	}

	// 2.2) Берём третью зону "mail.isHasMail"
	regionMailIsHasMailRaw, ok := lookup.Get("mail.isHasMail")
	if !ok {
		log.Fatal("area 'mail' not found")
	}

	regionMailIsHasMail := ocrclient.Region{
		X0: regionMailIsHasMailRaw.Zone.Min.X,
		Y0: regionMailIsHasMailRaw.Zone.Min.Y,
		X1: regionMailIsHasMailRaw.Zone.Max.X,
		Y1: regionMailIsHasMailRaw.Zone.Max.Y,
	}

	// 2.3) Берём четвёртую зону "alliance.state.isAllianceContributeButton"
	regionAllianceStateIsAllianceContributeButtonRaw, ok := lookup.Get("alliance.state.isAllianceContributeButton")
	if !ok {
		log.Fatal("area 'alliance.state.isAllianceContributeButton' not found")
	}

	regionAllianceStateIsAllianceContributeButton := ocrclient.Region{
		X0: regionAllianceStateIsAllianceContributeButtonRaw.Zone.Min.X,
		Y0: regionAllianceStateIsAllianceContributeButtonRaw.Zone.Min.Y,
		X1: regionAllianceStateIsAllianceContributeButtonRaw.Zone.Max.X,
		Y1: regionAllianceStateIsAllianceContributeButtonRaw.Zone.Max.Y,
	}

	// 3) Создаём ocr-клиент (лог можно передать nil, если не нужен)
	client := ocrclient.NewClient("RF8RC00M8MF", logger)

	// 4) Формируем FetchOCRRequest с регионом
	reqBody := ocrclient.FetchOCRRequest{
		DeviceID:  client.DeviceID,
		DebugName: "power_zone",
		Regions:   []ocrclient.Region{regionVIPIsNotify, regionGems, regionMailIsHasMail, regionAllianceStateIsAllianceContributeButton},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("marshal request: %v", err)
	}

	// 5) POST + замер времени
	url := client.ServiceURL + "/ocr"
	httpClient := &http.Client{Timeout: 20 * time.Second}
	start := time.Now()
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(payload))
	elapsed := time.Since(start)

	if err != nil {
		log.Fatalf("OCR request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("unexpected status %d: %s", resp.StatusCode, body)
	}

	// 6) Декодируем []OCRZone (все поля уже описаны в ocrclient)
	var zones []ocrclient.OCRZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		log.Fatalf("decode response: %v", err)
	}

	// 7) Печатаем результат и время
	for _, z := range zones {
		res := z.ToOCRResult()
		fmt.Printf("Text: %-20s  Score: %.2f  AvgColor: %-6s  BgColor: %-6s\n",
			res.Text, res.Score, res.AvgColor, res.BgColor)
	}
	fmt.Printf("Elapsed: %v\n", elapsed)
}
