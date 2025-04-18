package fsm

import (
	"log/slog"
)

func (g *GameFSM) Back() {
	if g.Current() == StateMail && g.previousState != "" {
		g.logger.Info("Returning from mail to previous state", slog.String("to", g.previousState))
		g.ForceTo(g.previousState)
	}
}
