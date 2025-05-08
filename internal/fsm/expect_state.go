package fsm

import (
	"log/slog"

	"github.com/samber/lo"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
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

	ocrTitle := gamerState.ScreenState.TitleFact
	ocrFamily := gamerState.ScreenState.IsMainCity

	g.logger.Debug("FSM: анализ состояния",
		slog.String("ocr_title", ocrTitle),
		slog.String("ocr_family", ocrFamily),
		slog.String("want", want),
	)

	// 0. Если семейство определено — используем только его группу
	switch {
	case vision.FuzzySubstringMatch(ocrFamily, "world", 1):
		g.logger.Debug("FSM: Семейство определено как ГОРОД (по ссылке на мир)",
			slog.String("ocr_family", ocrFamily),
			slog.String("want", want),
		)
		groupStates, ok := config.TitleToState["MainCity"]
		g.logger.Debug("FSM: Группа MainCity", slog.Any("groupStates", groupStates), slog.Bool("found", ok))
		if ok && lo.Contains(groupStates, want) {
			g.logger.Info("FSM: want найден в группе MainCity", slog.String("state", want))
			return want, nil
		}
		g.logger.Warn("FSM: want не найден в группе MainCity, возвращаем первый элемент группы", slog.String("state", groupStates[0]))
		return groupStates[0], nil
	case vision.FuzzySubstringMatch(ocrFamily, "city", 1):
		g.logger.Debug("FSM: Семейство определено как МИР (по ссылке на город)",
			slog.String("ocr_family", ocrFamily),
			slog.String("want", want),
		)
		groupStates, ok := config.TitleToState["World"]
		g.logger.Debug("FSM: Группа World", slog.Any("groupStates", groupStates), slog.Bool("found", ok))
		if ok && lo.Contains(groupStates, want) {
			g.logger.Info("FSM: want найден в группе World", slog.String("state", want))
			return want, nil
		}
		g.logger.Warn("FSM: want не найден в группе World, возвращаем первый элемент группы", slog.String("state", groupStates[0]))
		return groupStates[0], nil
	}

	// 1. Находим группу стейтов по заголовку (старый путь)
	groupStates, ok := getMatchedState(ocrTitle, 1)
	g.logger.Debug("FSM: Результат getMatchedState по ocrTitle",
		slog.String("ocr_title", ocrTitle),
		slog.Any("groupStates", groupStates),
		slog.Bool("found", ok),
	)
	if !ok {
		g.logger.Warn("FSM: Не удалось сопоставить заголовок ни с одной группой стейтов",
			slog.String("ocr_title", ocrTitle),
			slog.String("expected_state", want),
		)
		return want, nil
	}

	// 2. Ограничиваем группу по семейству (если можем)
	var filteredGroup []string
	switch {
	case vision.FuzzySubstringMatch(ocrFamily, "world", 1):
		filteredGroup = lo.Filter(groupStates, func(s string, _ int) bool {
			return s == state.StateMainCity
		})
		g.logger.Debug("FSM: Ограничение по семейству ГОРОД (MainCity)",
			slog.Any("filteredGroup", filteredGroup),
		)
	case vision.FuzzySubstringMatch(ocrFamily, "city", 1):
		filteredGroup = lo.Filter(groupStates, func(s string, _ int) bool {
			return s == state.StateWorld
		})
		g.logger.Debug("FSM: Ограничение по семейству МИР (World)",
			slog.Any("filteredGroup", filteredGroup),
		)
	default:
		filteredGroup = groupStates
		g.logger.Debug("FSM: Семейство не определено — используем всю группу",
			slog.Any("filteredGroup", filteredGroup),
		)
	}

	if lo.Contains(filteredGroup, want) {
		g.logger.Info("FSM: want найден в фильтрованной группе", slog.String("state", want))
		return want, nil
	}
	g.logger.Warn("FSM: want не найден в фильтрованной группе, возвращаем want", slog.String("state", want))
	return want, nil
}

func getMatchedState(title string, maxDistance int) ([]string, bool) {
	for key, states := range config.TitleToState {
		if vision.FuzzySubstringMatch(title, key, maxDistance) {
			return states, true
		}
	}
	return nil, false
}
