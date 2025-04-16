package device

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func (d *Device) findEmailOCR(ctx context.Context, email string) *domain.OCRResult {
	if cached, ok := d.getCachedEmailOCR(ctx, email); ok {
		d.Logger.Debug("📦 Email OCR из Redis", slog.String("email", email))
		return cached
	}

	zones, err := vision.WaitForText(ctx, d.ADB, []string{email}, time.Second, image.Rectangle{})
	if err != nil {
		d.Logger.Error("❌ Не удалось найти email на экране", slog.String("email", email), slog.Any("error", err))
		panic(fmt.Sprintf("WaitForText(%s) failed: %v", email, err))
	}

	d.setCachedEmailOCR(ctx, email, zones)
	return zones
}

func (d *Device) getCachedEmailOCR(ctx context.Context, email string) (*domain.OCRResult, bool) {
	key := fmt.Sprintf("bot:ocr:%s:email:%s", d.Name, email)
	val, err := d.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var result domain.OCRResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}

	return &result, true
}

func (d *Device) setCachedEmailOCR(ctx context.Context, email string, ocr *domain.OCRResult) {
	key := fmt.Sprintf("bot:ocr:%s:email:%s", d.Name, email)
	data, err := json.Marshal(ocr)
	if err != nil {
		d.Logger.Warn("❌ Не удалось сериализовать OCRResult", slog.Any("error", err))
		return
	}

	_ = d.rdb.Set(ctx, key, data, 12*time.Hour).Err()
}

func (d *Device) findNicknameOCR(ctx context.Context, nickname string) *domain.OCRResult {
	if cached, ok := d.getCachedNicknameOCR(ctx, nickname); ok {
		d.Logger.Debug("📦 Nickname OCR из Redis", slog.String("nickname", nickname))
		return cached
	}

	zones, err := vision.WaitForText(ctx, d.ADB, []string{nickname}, time.Second, image.Rectangle{})
	if err != nil {
		d.Logger.Error("❌ Не удалось найти nickname на экране", slog.String("nickname", nickname), slog.Any("error", err))
		panic(fmt.Sprintf("WaitForText(%s) failed: %v", nickname, err))
	}

	d.setCachedNicknameOCR(ctx, nickname, zones)
	return zones
}

func (d *Device) getCachedNicknameOCR(ctx context.Context, nickname string) (*domain.OCRResult, bool) {
	key := fmt.Sprintf("bot:ocr:%s:nickname:%s", d.Name, nickname)
	val, err := d.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var result domain.OCRResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}

	return &result, true
}

func (d *Device) setCachedNicknameOCR(ctx context.Context, nickname string, ocr *domain.OCRResult) {
	key := fmt.Sprintf("bot:ocr:%s:nickname:%s", d.Name, nickname)
	data, err := json.Marshal(ocr)
	if err != nil {
		d.Logger.Warn("❌ Не удалось сериализовать OCRResult", slog.Any("error", err))
		return
	}

	_ = d.rdb.Set(ctx, key, data, 12*time.Hour).Err()
}
