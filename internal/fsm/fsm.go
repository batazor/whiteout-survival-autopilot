package fsm

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	lpfsm "github.com/looplab/fsm"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type StateUpdateCallback interface {
	UpdateStateFromScreenshot(screen string)
}

// --------------------------------------------------------------------
// State Definitions: Each constant represents a game screen (state)
// --------------------------------------------------------------------
const (
	InitialState           = "initial"
	StateMainCity          = "main_city"
	StateActivityTriumph   = "activity_triumph"
	StateAllianceManage    = "alliance_manage"
	StateAllianceSettings  = "alliance_settings"
	StateAllianceHistory   = "alliance_history"
	StateAllianceList      = "alliance_list"
	StateAllianceVote      = "alliance_vote"
	StateAllianceRanking   = "alliance_ranking"
	StateEvents            = "events"
	StateProfile           = "profile"
	StateLeaderboard       = "leaderboard"
	StateSettings          = "settings"
	StateVIP               = "vip"
	StateChiefOrders       = "chief_orders"
	StateMail              = "mail"
	StateDawnMarket        = "dawn_market"
	StateExploration       = "exploration"
	StateExplorationButtle = "exploration_buttle"
)

var AllStates = []string{
	InitialState,
	StateMainCity,
	StateActivityTriumph,
	StateAllianceManage,
	StateAllianceSettings,
	StateAllianceHistory,
	StateAllianceList,
	StateAllianceVote,
	StateAllianceRanking,
	StateEvents,
	StateProfile,
	StateLeaderboard,
	StateSettings,
	StateVIP,
	StateChiefOrders,
	StateMail,
	StateDawnMarket,
	StateExploration,
	StateExplorationButtle,
}

type TransitionStep struct {
	Action string
	Wait   time.Duration
}

var transitionPaths = map[string]map[string][]TransitionStep{
	StateMainCity: {
		StateExploration:    {{Action: "to_exploration", Wait: 300 * time.Millisecond}},
		StateEvents:         {{Action: "to_events", Wait: 300 * time.Millisecond}},
		StateProfile:        {{Action: "to_profile", Wait: 300 * time.Millisecond}},
		StateLeaderboard:    {{Action: "to_leaderboard", Wait: 300 * time.Millisecond}},
		StateSettings:       {{Action: "to_settings", Wait: 300 * time.Millisecond}},
		StateVIP:            {{Action: "to_vip", Wait: 300 * time.Millisecond}},
		StateChiefOrders:    {{Action: "to_chief_orders", Wait: 300 * time.Millisecond}},
		StateMail:           {{Action: "to_mail", Wait: 300 * time.Millisecond}},
		StateDawnMarket:     {{Action: "to_dawn_market", Wait: 300 * time.Millisecond}},
		StateAllianceManage: {{Action: "to_alliance_manage", Wait: 300 * time.Millisecond}},
		StateAllianceSettings: {
			{Action: "to_alliance_manage", Wait: 300 * time.Millisecond},
			{Action: "to_alliance_settings", Wait: 300 * time.Millisecond},
		},
	},
	StateEvents: {
		StateActivityTriumph: {{Action: "to_activity_triumph", Wait: 300 * time.Millisecond}},
	},
	StateAllianceManage: {
		StateAllianceHistory:  {{Action: "to_alliance_history", Wait: 300 * time.Millisecond}},
		StateAllianceList:     {{Action: "to_alliance_list", Wait: 300 * time.Millisecond}},
		StateAllianceVote:     {{Action: "to_alliance_vote", Wait: 300 * time.Millisecond}},
		StateAllianceRanking:  {{Action: "to_alliance_ranking", Wait: 300 * time.Millisecond}},
		StateAllianceSettings: {{Action: "to_alliance_settings", Wait: 300 * time.Millisecond}},
	},
	StateExploration: {
		StateExplorationButtle: {{Action: "to_exploration_buttle", Wait: 300 * time.Millisecond}},
	},
}

type GameFSM struct {
	fsm           *lpfsm.FSM
	logger        *slog.Logger
	onStateChange func(state string)
	callback      StateUpdateCallback
	getState      func() *domain.State
	controller    interface {
		ClickRegion(name string, areas map[string]config.Region) error
	}
	areas map[string]config.Region
}

func NewGameFSM(logger *slog.Logger) *GameFSM {
	g := &GameFSM{logger: logger}

	transitions := lpfsm.Events{}
	callbacks := lpfsm.Callbacks{
		"enter_state": func(ctx context.Context, e *lpfsm.Event) {
			if g.logger != nil {
				g.logger.Info("FSM entered new state",
					slog.String("from", e.Src),
					slog.String("to", e.Dst),
					slog.String("event", e.Event),
				)
			}
			if g.onStateChange != nil {
				g.onStateChange(e.Dst)
			}
		},
	}

	g.fsm = lpfsm.NewFSM(InitialState, transitions, callbacks)
	return g
}

func (g *GameFSM) SetCallback(cb StateUpdateCallback) {
	g.callback = cb
}

func (g *GameFSM) SetStateGetter(getter func() *domain.State) {
	g.getState = getter
}

func (g *GameFSM) SetOnStateChange(f func(state string)) {
	g.onStateChange = f
}

func (g *GameFSM) SetController(ctrl interface {
	ClickRegion(string, map[string]config.Region) error
}, areas map[string]config.Region) {
	ValidateTransitionActions(areas)
	g.controller = ctrl
	g.areas = areas
}

func (g *GameFSM) Current() string {
	return g.fsm.Current()
}

func (g *GameFSM) ForceTo(target string) {
	prev := g.Current()

	if g.controller != nil && transitionPaths[prev] != nil {
		if steps, ok := transitionPaths[prev][target]; ok {
			g.logger.Info("FSM multi-step transition", slog.String("from", prev), slog.String("to", target), slog.Any("steps", steps))
			for _, step := range steps {
				err := g.controller.ClickRegion(step.Action, g.areas)
				if err != nil {
					g.logger.Error("Transition step failed", slog.String("action", step.Action), slog.Any("error", err))
					break
				}
				wait := step.Wait + time.Duration(rand.Intn(300)+100)*time.Millisecond
				g.logger.Debug("Waiting after action", slog.String("action", step.Action), slog.Duration("wait", wait))
				time.Sleep(wait)
			}
		} else {
			for from, targets := range transitionPaths {
				if subSteps, ok := targets[target]; ok {
					_ = g.tryTransitionVia(from, subSteps)
					break
				}
			}
		}
	}

	g.fsm.SetState(target)

	if g.logger != nil {
		g.logger.Warn("FSM forcefully moved to new state", slog.String("from", prev), slog.String("to", target))
	}

	if g.callback != nil {
		g.callback.UpdateStateFromScreenshot(target)
	}
}

func (g *GameFSM) tryTransitionVia(from string, steps []TransitionStep) error {
	g.logger.Info("Trying indirect transition", slog.String("via", from), slog.Any("steps", steps))
	for _, step := range steps {
		err := g.controller.ClickRegion(step.Action, g.areas)
		if err != nil {
			g.logger.Error("Indirect transition failed", slog.String("step", step.Action), slog.Any("error", err))
			return err
		}
		wait := step.Wait + time.Duration(rand.Intn(300)+100)*time.Millisecond
		time.Sleep(wait)
	}
	return nil
}

func ValidateTransitionActions(areas map[string]config.Region) {
	missing := make(map[string][]string)
	for from, targets := range transitionPaths {
		for to, steps := range targets {
			for _, step := range steps {
				if _, ok := areas[step.Action]; !ok {
					missing[from] = append(missing[from], fmt.Sprintf("%s → %s: '%s'", from, to, step.Action))
				}
			}
		}
	}
	if len(missing) > 0 {
		errMsg := "❌ Missing required region definitions in area.json:\n"
		for _, issues := range missing {
			for _, entry := range issues {
				errMsg += " - " + entry + "\n"
			}
		}
		panic(errMsg)
	}
}
