// Package ocrclient provides convenient methods to call the OCR HTTP service
// from your bot, including plain OCR, wait-for-text, image-find and region-based OCR.
package ocrclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// FindImageResponse is the response from /find_image.
type FindImageResponse struct {
	Found bool      `json:"found"`
	Boxes [][][]int `json:"boxes"`
}

// ToRects конвертирует каждую полигональную рамку в прямоугольник image.Rectangle.
func (r *FindImageResponse) ToRects() []image.Rectangle {
	rects := make([]image.Rectangle, 0, len(r.Boxes))
	for _, poly := range r.Boxes {
		if len(poly) == 0 {
			continue
		}
		// инициализируем габариты первыми точками
		minX, minY := poly[0][0], poly[0][1]
		maxX, maxY := minX, minY
		for _, pt := range poly {
			x, y := pt[0], pt[1]
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
		rects = append(rects, image.Rect(minX, minY, maxX, maxY))
	}
	return rects
}

// FetchOCRRequest is the JSON payload for the /ocr endpoint.
type FetchOCRRequest struct {
	DeviceID  string   `json:"device_id,omitempty"`
	DebugName string   `json:"debug_name,omitempty"`
	Regions   []Region `json:"regions,omitempty"`
}

// FetchOCR performs a one‐shot OCR by POSTing JSON to the /ocr endpoint,
// using a 20s timeout on the HTTP request.
func (c *Client) FetchOCR(debugName string, regions []Region) (domain.OCRResults, error) {
	c.Logger.Info("🖼️  Fetching OCR",
		"device_id", c.DeviceID,
		"debug_name", debugName,
	)

	// prepare JSON body
	reqBody := FetchOCRRequest{
		DeviceID:  c.DeviceID,
		DebugName: debugName,
		Regions:   regions,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal /ocr payload: %w", err)
	}

	// build POST request
	req, err := http.NewRequest("POST", c.ServiceURL+"/ocr", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request /ocr: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 3) Выполняем через c.HTTP — с автоматическими retry и таймаутом
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post /ocr: %w", err)
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	// decode OCRZone slice
	var zones []OCRZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return nil, fmt.Errorf("decode /ocr json: %w", err)
	}

	// convert to domain.OCRResults
	results := make(domain.OCRResults, len(zones))
	for i, z := range zones {
		results[i] = z.ToOCRResult()
	}
	return results, nil
}

// WaitForTextRequest — payload для /wait_for_text.
type WaitForTextRequest struct {
	StopWords []string `json:"stop_words"`
	DeviceID  string   `json:"device_id,omitempty"`
	Timeout   float64  `json:"timeout"`
	Interval  float64  `json:"interval"`
}

// WaitForText polls /wait_for_text until one of stopWords appears or timeout elapses.
// timeout and interval are now time.Duration.
func (c *Client) WaitForText(stopWords []string, timeout, interval time.Duration, debugName string) (domain.OCRResults, error) {
	c.Logger.Info("🖼️ Waiting for text",
		"device_id", c.DeviceID,
		"debug_name", debugName,
		"stop_words", stopWords,
		"timeout", timeout,
		"interval", interval,
	)

	// 1) Подготовка тела запроса
	reqBody := WaitForTextRequest{
		StopWords: stopWords,
		DeviceID:  c.DeviceID,
		Timeout:   timeout.Seconds(),
		Interval:  interval.Seconds(),
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal /wait_for_text payload: %w", err)
	}

	// 2) Формируем URL с debug_name
	u, err := url.Parse(c.ServiceURL + "/wait_for_text")
	if err != nil {
		return nil, fmt.Errorf("parse wait_for_text URL: %w", err)
	}
	if debugName != "" {
		q := u.Query()
		q.Set("debug_name", debugName)
		u.RawQuery = q.Encode()
	}

	// 3) Собираем HTTP-запрос
	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("new request /wait_for_text: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 4) Выполняем через c.HTTP (retry + timeout)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post /wait_for_text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, data)
	}

	var zones []OCRZone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return nil, fmt.Errorf("decode /wait_for_text json: %w", err)
	}

	// Преобразуем каждый OCRZone в OCRResult
	results := make(domain.OCRResults, len(zones))
	for i, z := range zones {
		results[i] = z.ToOCRResult()
	}
	return results, nil
}

// FindImageRequest is the JSON payload for the /find_image endpoint.
type FindImageRequest struct {
	ImageName string   `json:"image_name"`
	DeviceID  string   `json:"device_id,omitempty"`
	Threshold float64  `json:"threshold"`
	DebugName string   `json:"debug_name,omitempty"`
	Regions   []Region `json:"regions,omitempty"`
}

// FindImage searches for all occurrences of imageName in the screen.
// It uses a 30s timeout on the HTTP request.
func (c *Client) FindImage(imageName string, threshold float64, debugName string) (*FindImageResponse, error) {
	c.Logger.Info("🖼️  Finding image",
		"device_id", c.DeviceID,
		"image_name", imageName,
		"threshold", threshold,
		"debug_name", debugName,
	)

	// 1) Подготовка запроса
	reqBody := FindImageRequest{
		ImageName: imageName,
		DeviceID:  c.DeviceID,
		Threshold: threshold,
		DebugName: debugName,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal find_image payload: %w", err)
	}

	// 2) Формируем HTTP-запрос
	url := fmt.Sprintf("%s/find_image", c.ServiceURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("new request /find_image: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 3) Выполняем через c.HTTP (retry + timeout)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post /find_image: %w", err)
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
