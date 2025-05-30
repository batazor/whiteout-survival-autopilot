package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

// тот же Region для JSON
type Region struct {
	X0 int `json:"x0"`
	Y0 int `json:"y0"`
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
}

type OCRZone struct {
	Box      [][]int `json:"box"`
	Text     string  `json:"text"`
	Score    float64 `json:"score"`
	AvgColor string  `json:"avg_color"`
	BgColor  string  `json:"bg_color"`
	Crop     string  `json:"crop,omitempty"`
}

// FindImageResponse соответствует JSON-ответу /find_image
type FindImageResponse struct {
	Found bool      `json:"found"`
	Boxes [][][]int `json:"boxes"` // каждый box — [[x1,y1],[x2,y2],[x3,y3],[x4,y4]]
}

// fetchOCR делает GET /ocr?device_id=&debug_name=
func fetchOCR(serviceURL, deviceID, debugName string) ([]OCRZone, error) {
	u, err := url.Parse(serviceURL + "/ocr")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if deviceID != "" {
		q.Set("device_id", deviceID)
	}
	if debugName != "" {
		q.Set("debug_name", debugName)
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("http get /ocr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	var zones []OCRZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return nil, fmt.Errorf("decode /ocr json: %w", err)
	}
	return zones, nil
}

// fetchWaitForText делает POST /wait_for_text?debug_name=
func fetchWaitForText(serviceURL, deviceID, debugName string, stopWords []string, timeout, interval float64) ([]OCRZone, error) {
	// тело запроса
	body := map[string]interface{}{
		"stop_words": stopWords,
		"device_id":  deviceID,
		"timeout":    timeout,
		"interval":   interval,
	}
	b, _ := json.Marshal(body)

	// строим URL с debug_name
	u, _ := url.Parse(serviceURL + "/wait_for_text")
	if debugName != "" {
		q := u.Query()
		q.Set("debug_name", debugName)
		u.RawQuery = q.Encode()
	}

	resp, err := http.Post(u.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("http post /wait_for_text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	var zones []OCRZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return nil, fmt.Errorf("decode /wait_for_text json: %w", err)
	}
	return zones, nil
}

func callExampleOCR() {
	fmt.Println("=== Example OCR ===")
	svc := "http://localhost:8000"
	serial := "RF8RC00M8MF"
	debug := "ocr_debug.png"
	zones, err := fetchOCR(svc, serial, debug)
	if err != nil {
		fmt.Println("ERROR fetchOCR:", err)
		return
	}
	for i, z := range zones {
		fmt.Printf("[%d] box=%v text=%q score=%.2f avg=%s bg=%s\n",
			i, z.Box, z.Text, z.Score, z.AvgColor, z.BgColor)
	}
}

func callExampleWait() {
	fmt.Println("=== Example WaitForText ===")
	svc := "http://localhost:8000"
	serial := "RF8RC00M8MF"
	debug := "wait_debug.png"
	stopWords := []string{"Claim"}
	zones, err := fetchWaitForText(svc, serial, debug, stopWords, 15.0, 1.0)
	if err != nil {
		fmt.Println("ERROR fetchWaitForText:", err)
		return
	}
	if len(zones) == 0 {
		fmt.Println("Timeout reached, no stop words found.")
		return
	}
	for i, z := range zones {
		fmt.Printf("[%d] FOUND %q at box=%v score=%.2f avg=%s bg=%s\n",
			i, z.Text, z.Box, z.Score, z.AvgColor, z.BgColor)
	}
}

// fetchFindImage делает GET /find_image?image_name=<name>&device_id=<serial>&threshold=<t>&debug_name=<fn>
func fetchFindImage(serviceURL, deviceID, imageName string, threshold float64, debugName string) (*FindImageResponse, error) {
	u, err := url.Parse(serviceURL + "/find_image")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("image_name", imageName)
	if deviceID != "" {
		q.Set("device_id", deviceID)
	}
	if threshold > 0 {
		q.Set("threshold", fmt.Sprintf("%.2f", threshold))
	}
	if debugName != "" {
		q.Set("debug_name", debugName)
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("http get /find_image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	var result FindImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode /find_image json: %w", err)
	}
	return &result, nil
}

func callExampleFind() {
	fmt.Println("=== Example Find Image ===")
	svc := "http://localhost:8000"
	serial := "RF8RC00M8MF"                     // ваш ADB serial или пусто
	imageName := "alliance.state.isNeedSupport" // имя файла ../references/screenshots/handshake.png
	threshold := 0.8
	debug := "find_debug.png"

	res, err := fetchFindImage(svc, serial, imageName, threshold, debug)
	if err != nil {
		fmt.Println("ERROR fetchFindImage:", err)
		return
	}
	if !res.Found {
		fmt.Println("Image not found on screen.")
		return
	}
	fmt.Printf("Found %d occurrences of '%s':\n", len(res.Boxes), imageName)
	for i, box := range res.Boxes {
		fmt.Printf(" #%d box=%v\n", i, box)
	}
}

func callExampleOCRByZoneName() {
	// 1) Загружаем area.json
	lookup, err := config.LoadAreaReferences("references/area.json")
	if err != nil {
		panic(fmt.Errorf("load area references: %w", err))
	}

	// 2) Получаем *BBox по имени
	bbox, err := lookup.GetRegionByName("screenState.titleFact")
	if err != nil {
		fmt.Println("Region not found:", err)
		return
	}

	// 3) ToRectangle → image.Rectangle
	rect := bbox.ToRectangle() // здесь используется ваш метод

	// 4) Преобразуем в наш Region
	region := Region{
		X0: rect.Min.X,
		Y0: rect.Min.Y,
		X1: rect.Max.X,
		Y1: rect.Max.Y,
	}

	// 5) Формируем и шлём POST /ocr
	payload := struct {
		DeviceID  string   `json:"device_id,omitempty"`
		DebugName string   `json:"debug_name,omitempty"`
		Regions   []Region `json:"regions,omitempty"`
	}{
		DeviceID:  "RF8RC00M8MF",
		DebugName: "titleFact_debug.png",
		Regions:   []Region{region},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("marshal error:", err)
		return
	}

	resp, err := http.Post("http://localhost:8000/ocr", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println("ERROR http post /ocr:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		fmt.Printf("bad status %d: %s\n", resp.StatusCode, data)
		return
	}

	// 6) Декодируем и выводим
	var zones []OCRZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		fmt.Println("ERROR decode JSON:", err)
		return
	}
	fmt.Printf("Found %d OCR zones in screenState.titleFact:\n", len(zones))
	for i, z := range zones {
		fmt.Printf(
			" [%d] box=%v text=%q score=%.2f avg=%s bg=%s\n",
			i, z.Box, z.Text, z.Score, z.AvgColor, z.BgColor,
		)
	}
}

func main() {
	//callExampleOCR()
	//callExampleWait()
	//callExampleFind()
	callExampleOCRByZoneName()
}
