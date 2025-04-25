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
						slog.String("action", step.Action),
						slog.String("trigger", step.Trigger),
						slog.Any("error", err),
					)
					panic("Trigger evaluation failed")
				}
				if !ok {
					g.logger.Info("Trigger condition not met, skipping step",
						slog.String("action", step.Action),
						slog.String("trigger", step.Trigger),
					)

					return EventNotActive
				}
			}

			if _, ok := g.lookup.Get(step.Action); !ok {
				panic(fmt.Sprintf("❌ Region '%s' not found in area.json", step.Action))
			}

			g.logger.Info("Clicking region", slog.String("action", step.Action))

			if err := g.adb.ClickRegion(step.Action, g.lookup); err != nil {
				panic(fmt.Sprintf("❌ ADB click failed for action '%s': %v", step.Action, err))
			}

			wait := step.Wait + time.Duration(rand.Intn(300)+700)*time.Millisecond
			g.logger.Info("Waiting after action", slog.String("action", step.Action), slog.Duration("wait", wait))
			time.Sleep(wait)

			if g.callback != nil && len(path) > 0 && i+1 < len(path) {
				state := path[i+1]
				g.fsm.SetState(state)
				g.logger.Info("FSM intermediate state updated",
					slog.String("state", state),
					slog.String("step", step.Action),
				)
				g.callback.UpdateStateFromScreenshot(state)

				if updateStateFromScreen != nil {
					updateStateFromScreen(context.Background(), state, fmt.Sprintf("out/bot_%s_%s.png", g.gamerState.Nickname, target))
				}
			}
		}
	}

	// Try using the FSM event system first if possible
	eventName := fmt.Sprintf("%s_to_%s", prev, target)
	if err := g.fsm.Event(context.Background(), eventName); err != nil {
		// If the event isn't defined, fall back to direct state change
		g.fsm.SetState(target)
		g.logger.Warn("FSM forcefully moved to new state",
			slog.String("from", prev),
			slog.String("to", target),
		)
	}

	return nil
}
