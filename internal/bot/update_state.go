package bot

import (
	"context"
	"log/slog"
)

func (b *Bot) updateStateFromScreen(ctx context.Context, screen string, filename string) {
	if _, err := b.Device.ADB.Screenshot(filename); err != nil {
		b.Logger.Warn("⚠️ Не удалось сделать скриншот", slog.String("screen", screen), slog.Any("error", err))
		return
	}

	rules := b.Rules[screen]
	newState, err := b.Device.Executor.Analyzer().AnalyzeAndUpdateState(filename, b.Gamer, rules)
	if err != nil {
		b.Logger.Warn("⚠️ Ошибка анализа экрана", slog.String("screen", screen), slog.Any("error", err))
		return
	}

	*b.Gamer = *newState
	b.Logger.Info("📥 Состояние обновлено", slog.String("screen", screen))

	if saveErr := b.Repo.SaveGamer(ctx, newState); saveErr != nil {
		b.Logger.Error("❌ Не удалось сохранить state.yaml", slog.Any("error", saveErr))
	} else {
		b.Logger.Info("💾 Состояние игрока сохранено в state.yaml")
	}
}
