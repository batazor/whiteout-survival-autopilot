package device

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func (d *Device) NextGamer(profileIdx, gamerIdx int) {
	ctx := context.Background()

	d.activeProfileIdx = profileIdx
	d.activeGamerIdx = gamerIdx

	profile := d.Profiles[profileIdx]
	gamer := &profile.Gamer[gamerIdx]

	d.Logger.Info(ctx, "🎮 Переключение на другого игрока в текущем профиле",
		slog.String("email", profile.Email),
		slog.String("nickname", gamer.Nickname),
		slog.Int("id", gamer.ID),
	)

	// Устанавливаем нового игрока в FSM
	d.FSM.SetCallback(gamer)

	// 🔁 Навигация: переходим к экрану выбора аккаунта Google
	d.Logger.Info(ctx, "➡️ Переход в экран выбора игрока")
	d.FSM.ForceTo(ctx, fsm.StateChiefCharacters)

	// ждем nickname
	gamerZones, _ := vision.WaitForText(ctx, d.ADB, []string{gamer.Nickname}, time.Second, image.Rectangle{})

	d.Logger.Info(ctx, "🟢 Клик по nickname игрока", slog.String("text", gamerZones.Text))
	if err := d.ADB.ClickOCRResult(gamerZones); err != nil {
		d.Logger.Error(ctx, "❌ Не удалось кликнуть по nickname аккаунту", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(nickname:%s) failed: %v", gamer.Nickname, err))
	}

	time.Sleep(2 * time.Second)

	d.Logger.Info(ctx, "🟢 Клик по кнопке подтверждения", slog.String("region", "character_change_confirm"))
	if err := d.ADB.ClickRegion("character_change_confirm", d.AreaLookup); err != nil {
		d.Logger.Error(ctx, "❌ Не удалось кликнуть по character_change_confirm", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(character_change_confirm) failed: %v", err))
	}

	// Проверка стартовых баннеров
	err := d.handleEntryScreens(ctx)
	if err != nil {
		d.Logger.Error(ctx, "❌ Не удалось обработать стартовые баннеры", slog.Any("err", err))
		panic(fmt.Sprintf("handleEntryScreens() failed: %v", err))
	}

	d.Logger.Info(ctx, "✅ Вход выполнен, переход в Main City")
	d.Logger.Info(ctx, "🔧 Инициализация FSM")
	d.FSM = fsm.NewGame(ctx, d.Logger, d.ADB, d.AreaLookup)
}
