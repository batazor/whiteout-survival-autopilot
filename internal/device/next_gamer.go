package device

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

func (d *Device) NextGamer(profileIdx, gamerIdx int) {
	// Инициируем контекст и span
	ctx := context.Background()
	tracer := otel.Tracer("device")
	ctx, span := tracer.Start(ctx, "NextGamer")
	defer span.End()

	// Достаём traceID для логов
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	d.activeProfileIdx = profileIdx
	d.activeGamerIdx = gamerIdx

	profile := d.Profiles[profileIdx]
	gamer := &profile.Gamer[gamerIdx]

	d.Logger.Info("🎮 Переключение на другого игрока в текущем профиле",
		slog.String("email", profile.Email),
		slog.String("nickname", gamer.Nickname),
		slog.Int("id", gamer.ID),
		slog.String("trace_id", traceID),
	)

	// Устанавливаем нового игрока в FSM
	d.FSM.SetCallback(gamer)

	// 🔁 Навигация: переходим к экрану выбора аккаунта
	d.Logger.Info("➡️ Переход в экран выбора игрока",
		slog.String("trace_id", traceID),
	)
	d.FSM.ForceTo(state.StateChiefCharacters, nil)

	// 🕒 Ждём, чтобы не было конфликта с другими процессами
	time.Sleep(2 * time.Second)

	// ========== 1️⃣ Делаем единый full-screen OCR ==========
	fullOCR, fullErr := d.OCRClient.FetchOCR("", nil) // debugName можно опустить
	if fullErr != nil {
		d.Logger.Error("❌ Full OCR failed", slog.Any("error", fullErr))
		panic(fmt.Sprintf("ocrClient.FetchOCR() failed: %v", fullErr))
	}

	// Ждём nickname
	var gamerZone *domain.OCRResult
	for _, zone := range fullOCR {
		if strings.Contains(zone.Text, gamer.Nickname) {
			gamerZone = &zone
			break
		}
	}

	d.Logger.Info("🟢 Клик по nickname игрока",
		slog.String("text", gamerZone.Text),
		slog.String("trace_id", traceID),
	)
	if err := d.ADB.ClickOCRResult(gamerZone); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по nickname аккаунту",
			slog.Any("err", err),
			slog.String("trace_id", traceID),
		)
		panic(fmt.Sprintf("ClickRegion(nickname:%s) failed: %v", gamer.Nickname, err))
	}

	time.Sleep(2 * time.Second)

	d.Logger.Info("🟢 Клик по кнопке подтверждения",
		slog.String("region", "character_change_confirm"),
		slog.String("trace_id", traceID),
	)
	if err := d.ADB.ClickRegion("character_change_confirm", d.AreaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по character_change_confirm",
			slog.Any("err", err),
			slog.String("trace_id", traceID),
		)
		panic(fmt.Sprintf("ClickRegion(character_change_confirm) failed: %v", err))
	}

	// Проверка стартовых баннеров
	err := d.handleEntryScreens(ctx)
	if err != nil {
		d.Logger.Error("❌ Не удалось обработать стартовые баннеры",
			slog.Any("err", err),
			slog.String("trace_id", traceID),
		)
		panic(fmt.Sprintf("handleEntryScreens() failed: %v", err))
	}

	d.Logger.Info("✅ Вход выполнен, переход в Main City",
		slog.String("trace_id", traceID),
	)
	d.Logger.Info("🔧 Инициализация FSM",
		slog.String("trace_id", traceID),
	)
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.AreaLookup, d.triggerEvaluator, d.ActiveGamer(), d.OCRClient)
}
