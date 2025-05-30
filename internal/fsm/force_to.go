package fsm

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

var (
	EventNotActive = fmt.Errorf("event not active")
)

func (g *GameFSM) ForceTo(target string, updateStateFromScreen func(ctx context.Context, screen string, filename string)) error {
	prev := g.Current()

	// Save the previous state (before changing it)
	g.previousState = prev

	if prev == target {
		g.logger.Debug("FSM already at target state, skipping", slog.String("state", target))
		return nil
	}

	var steps []TransitionStep
	var path = []string{prev, target}
	found := false

	if g.adb != nil {
		steps, found = transitionPaths[prev][target]
		if !found {
			path = g.FindPath(prev, target)
			if len(path) > 1 {
				g.logger.Warn("FSM path generated dynamically", slog.Any("path", path))
				steps = g.pathToSteps(path)
				g.logAutoPath(path)
			} else {
				panic(fmt.Sprintf("❌ FSM: no path found from '%s' to '%s'", prev, target))
			}
		}

		for i, step := range steps {
			// Проверка Trigger (CEL)
			if step.Trigger != "" {
				ok, err := g.triggerEvaluator.EvaluateTrigger(step.Trigger, g.gamerState)
				if err != nil {
					g.logger.Error("Trigger evaluation failed",
						slog.String("trigger", step.Trigger),
						slog.Any("error", err),
					)
					panic("Trigger evaluation failed")
				}
				if !ok {
					g.logger.Info("Trigger condition not met, skipping step",
						slog.String("trigger", step.Trigger),
					)

					return EventNotActive
				}
			}

			// Проверка условий для клика
			if step.Click != "" {
				if _, ok := g.lookup.Get(step.Click); !ok {
					panic(fmt.Sprintf("❌ Region '%s' not found in area.json", step.Click))
				}

				g.logger.Info("Clicking region", slog.String("click", step.Click))

				if err := g.adb.ClickRegion(step.Click, g.lookup); err != nil {
					panic(fmt.Sprintf("❌ ADB click failed for action '%s': %v", step.Click, err))
				}
			}

			// Проверка условий для свайпа
			if step.Swipe != nil {
				g.logger.Info("Swiping",
					slog.Int("x1", step.Swipe.X1), slog.Int("y1", step.Swipe.Y1),
					slog.Int("x2", step.Swipe.X2), slog.Int("y2", step.Swipe.Y2),
					slog.Duration("wait", step.Wait),
				)

				if err := g.adb.Swipe(step.Swipe.X1, step.Swipe.Y1, step.Swipe.X2, step.Swipe.Y2, step.Wait); err != nil {
					panic(fmt.Sprintf("❌ ADB swipe failed for action '%s': %v", step.Swipe, err))
				}
			}

			// СКИП проверки стейта если это свайп
			if step.Swipe != nil {
				time.Sleep(step.Wait)
				continue
			}

			wait := step.Wait + time.Duration(rand.Intn(300)+700)*time.Millisecond
			g.logger.Info("Waiting after action", slog.String("click", step.Click), slog.Duration("wait", wait))
			time.Sleep(wait)

			expected := target
			if i+1 < len(path) {
				expected = path[i+1]
			}

			actual, errCheckState := g.ExpectState(expected)
			if errCheckState != nil {
				g.logger.Error("❌ Ошибка при проверке состояния после действия",
					slog.String("click", step.Click),
					slog.String("expected", expected),
					slog.String("actual", actual),
					slog.Any("error", errCheckState),
				)
				return errCheckState
			}

			if actual != expected {
				g.logger.Warn("⚠️ Обнаружено несоответствие состояния после действия",
					slog.String("click", step.Click),
					slog.String("expected", expected),
					slog.String("actual", actual),
				)

				// фиксируем актуальный стейт сразу в FSM и в стейте игрока!
				g.fsm.SetState(actual)
				g.gamerState.ScreenState.CurrentState = actual

				// пробуем построить путь к цели из текущего положения
				return g.ForceTo(target, updateStateFromScreen)
			}

			// Успешный шаг: синхронизируем FSM и состояние игрока
			g.fsm.SetState(actual)
			g.gamerState.ScreenState.CurrentState = actual

			// --- callback & скриншот -----------------------------------------------
			if g.callback != nil {
				if updateStateFromScreen != nil {
					updateStateFromScreen(
						context.Background(),
						actual,
						fmt.Sprintf(
							"out/bot_%s_%s.png",
							g.gamerState.Nickname,
							target,
						),
					)
				}

				next := target
				if i+1 < len(path) {
					next = path[i+1]
				}
				g.logger.Info("FSM state confirmed, next planned",
					slog.String("current", actual),
					slog.String("next", next),
					slog.String("step", step.Click),
				)
			}
		}
	}

	// финальная синхронизация
	eventName := fmt.Sprintf("%s_to_%s", prev, target)
	if err := g.fsm.Event(context.Background(), eventName); err != nil {
		// Если эвент не определён, форсируем смену состояния везде!
		g.fsm.SetState(target)
		g.logger.Warn("FSM forcefully moved to new state",
			slog.String("from", prev),
			slog.String("to", target),
		)
	}

	// В любом случае, после FSM-перехода (или ручного SetState) — синхронизируем gamerState:
	g.gamerState.ScreenState.CurrentState = target

	return nil
}
