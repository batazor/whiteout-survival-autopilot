package fsm

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

func (g *GameFSM) ForceTo(target string) {
	prev := g.Current()

	// Save the previous state (before changing it)
	g.previousState = prev

	if prev == target {
		g.logger.Debug("FSM already at target state, skipping", slog.String("state", target))
		return
	}

	var steps []TransitionStep
	found := false

	if g.adb != nil {
		steps, found = transitionPaths[prev][target]
		if !found {
			path := g.FindPath(prev, target)
			if len(path) > 1 {
				g.logger.Warn("FSM path generated dynamically", slog.Any("path", path))
				steps = g.pathToSteps(path)
				g.logAutoPath(path)
			} else {
				panic(fmt.Sprintf("❌ FSM: no path found from '%s' to '%s'", prev, target))
			}
		}

		for _, step := range steps {
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

	if g.callback != nil {
		g.callback.UpdateStateFromScreenshot(target)
	}
}
