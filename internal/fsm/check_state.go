package fsm

import (
	"log/slog"

	"github.com/samber/lo"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

// ExpectState выполняет проверку текущего состояния экрана.
// Если заголовок соответствует известному экрану, но стейт не совпадает — возвращает тот, что ожидается.
func (g *GameFSM) ExpectState(want string) (string, error) {
	screenshotPath := "out/check_state.png"

	_, err := g.adb.Screenshot(screenshotPath)
	if err != nil {
		g.logger.Error("FSM: failed to take screenshot",
			slog.String("action", "expect_state"),
			slog.String("state", want),
			slog.Any("error", err),
		)
		return "", err
	}

	gamerState, err := g.analyzer.AnalyzeAndUpdateState(
		screenshotPath, g.gamerState, g.rulesCheckState["default"], nil,
	)
	if err != nil {
		g.logger.Error("❌ Не удалось проанализировать экран",
			slog.String("action", "expect_state"),
			slog.String("state", want),
			slog.Any("error", err),
		)
		return "", err
	}

	actualTitle := gamerState.ScreenState.TitleFact

	// 1. Находим группу стейтов по заголовку (нечётко)
	groupStates, ok := getMatchedState(actualTitle, 1)
	if !ok {
		g.logger.Warn("⚠️ Не удалось сопоставить заголовок ни с одной группой стейтов",
			slog.String("ocr_title", actualTitle),
			slog.String("expected_state", want),
		)
		return want, nil
	}

	// 2. Проверяем, входит ли want в найденную группу
	if lo.Contains(groupStates, want) {
		g.logger.Info("✅ State confirmed in matched group",
			slog.String("ocr_title", actualTitle),
			slog.String("detected_state", want),
		)
		return want, nil
	}

	// 3. Если не входит — возвращаем первый из списка группы
	g.logger.Warn("⚠️ State doesn't match group, returning first from group",
		slog.String("ocr_title", actualTitle),
		slog.String("expected_state", want),
		slog.String("corrected_state", groupStates[0]),
	)
	return groupStates[0], nil
}

func getMatchedState(title string, maxDistance int) ([]string, bool) {
	for key, states := range config.TitleToState {
		if vision.FuzzySubstringMatch(title, key, maxDistance) {
			return states, true
		}
	}

	return nil, false
}
