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

	d.Logger.Info("🎮 Переключение на другого игрока в текущем профиле",
		slog.String("email", profile.Email),
		slog.String("nickname", gamer.Nickname),
		slog.Int("id", gamer.ID),
	)

	// Устанавливаем нового игрока в FSM
	d.FSM.SetCallback(gamer)

	// 🔁 Навигация: переходим к экрану выбора аккаунта Google
	d.Logger.Info("➡️ Переход в экран выбора игрока")
	d.FSM.ForceTo(fsm.StateChiefCharacters)

	// ждем nickname
	gamerZones, _ := vision.WaitForText(ctx, d.ADB, []string{gamer.Nickname}, time.Second, image.Rectangle{})

	d.Logger.Info("🟢 Клик по nickname игрока", slog.String("text", gamerZones.Text))
	if err := d.ADB.ClickOCRResult(gamerZones); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по nickname аккаунту", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(nickname:%s) failed: %v", gamer.Nickname, err))
	}

	time.Sleep(2 * time.Second)

	d.Logger.Info("🟢 Клик по кнопке подтверждения", slog.String("region", "character_change_confirm"))
	if err := d.ADB.ClickRegion("character_change_confirm", d.areaLookup); err != nil {
		d.Logger.Error("❌ Не удалось кликнуть по character_change_confirm", slog.Any("err", err))
		panic(fmt.Sprintf("ClickRegion(character_change_confirm) failed: %v", err))
	}

	// Проверка на страницу - добро пожаловать
	newCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	resp, _ := vision.WaitForText(newCtx, d.ADB, []string{"Welcome"}, time.Second, image.Rectangle{})

	if resp != nil {
		d.Logger.Info("🟢 Клик по кнопке Welcome Back", slog.String("region", "welcome_back_continue_button"))
		if err := d.ADB.ClickRegion("welcome_back_continue_button", d.areaLookup); err != nil {
			d.Logger.Error("❌ Не удалось кликнуть по welcome_back_continue_button", slog.Any("err", err))
			panic(fmt.Sprintf("ClickRegion(welcome_back_continue_button) failed: %v", err))
		}
	}

	d.Logger.Info("✅ Вход выполнен, переход в Main City")
	d.Logger.Info("🔧 Инициализация FSM")
	d.FSM = fsm.NewGame(d.Logger, d.ADB, d.areaLookup)
}
