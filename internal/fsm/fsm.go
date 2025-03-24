package fsm

import (
	"context"
	"fmt"
	"log/slog"

	lpfsm "github.com/looplab/fsm"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/utils"
)

type StateUpdateCallback interface {
	UpdateStateFromScreenshot(screen string)
}

// --------------------------------------------------------------------
// State Definitions: Each constant represents a game screen (state)
// --------------------------------------------------------------------
const (
	StateMainCity         = "main_city"
	StateActivityTriumph  = "activity_triumph"
	StateAllianceManage   = "alliance_manage"
	StateAllianceSettings = "alliance_settings" // 🆕 добавлено
	StateAllianceHistory  = "alliance_history"
	StateAllianceList     = "alliance_list"
	StateAllianceVote     = "alliance_vote"
	StateAllianceRanking  = "alliance_ranking"
	StateEvents           = "events"
	StateProfile          = "profile"
	StateLeaderboard      = "leaderboard"
	StateSettings         = "settings"
	StateVIP              = "vip"
	StateChiefOrders      = "chief_orders"
	StateMail             = "mail"
	StateDawnMarket       = "dawn_market"
	StateExploration      = "exploration"
)

// --------------------------------------------------------------------
// Event Definitions
// --------------------------------------------------------------------
type Event string

const (
	EventGoToAllianceManage   Event = "to_alliance_manage"
	EventGoToAllianceSettings Event = "to_alliance_settings" // 🆕 добавлено
	EventGoToEvents           Event = "to_events"
	EventGoToProfile          Event = "to_profile"
	EventGoToLeaderboard      Event = "to_leaderboard"
	EventGoToSettings         Event = "to_settings"
	EventGoToVIP              Event = "to_vip"
	EventGoToChiefOrders      Event = "to_chief_orders"
	EventGoToMail             Event = "to_mail"
	EventGoToDawnMarket       Event = "to_dawn_market"
	EventGoToExploration      Event = "to_exploration"
	EventGoToActivityTriumph  Event = "to_activity_triumph"
	EventGoToAllianceHistory  Event = "to_alliance_history"
	EventGoToAllianceList     Event = "to_alliance_list"
	EventGoToAllianceVote     Event = "to_alliance_vote"
	EventGoToAllianceRanking  Event = "to_alliance_ranking"
	EventBack                 Event = "back"
)

type GameFSM struct {
	fsm           *lpfsm.FSM
	logger        *slog.Logger
	onStateChange func(state string)
	callback      StateUpdateCallback
	getState      func() *domain.State
}

func NewGameFSM(logger *slog.Logger) *GameFSM {
	g := &GameFSM{logger: logger}

	transitions := lpfsm.Events{
		{Name: string(EventGoToAllianceManage), Src: []string{StateMainCity}, Dst: StateAllianceManage},
		{Name: string(EventGoToAllianceSettings), Src: []string{StateAllianceManage}, Dst: StateAllianceSettings}, // 🆕
		{Name: string(EventGoToEvents), Src: []string{StateMainCity}, Dst: StateEvents},
		{Name: string(EventGoToProfile), Src: []string{StateMainCity}, Dst: StateProfile},
		{Name: string(EventGoToLeaderboard), Src: []string{StateMainCity}, Dst: StateLeaderboard},
		{Name: string(EventGoToSettings), Src: []string{StateMainCity}, Dst: StateSettings},
		{Name: string(EventGoToVIP), Src: []string{StateMainCity}, Dst: StateVIP},
		{Name: string(EventGoToChiefOrders), Src: []string{StateMainCity}, Dst: StateChiefOrders},
		{Name: string(EventGoToMail), Src: []string{StateMainCity}, Dst: StateMail},
		{Name: string(EventGoToDawnMarket), Src: []string{StateMainCity}, Dst: StateDawnMarket},
		{Name: string(EventGoToExploration), Src: []string{StateMainCity}, Dst: StateExploration},
		{Name: string(EventGoToActivityTriumph), Src: []string{StateEvents}, Dst: StateActivityTriumph},
		{Name: string(EventGoToAllianceHistory), Src: []string{StateAllianceManage}, Dst: StateAllianceHistory},
		{Name: string(EventGoToAllianceList), Src: []string{StateAllianceManage}, Dst: StateAllianceList},
		{Name: string(EventGoToAllianceVote), Src: []string{StateAllianceManage}, Dst: StateAllianceVote},
		{Name: string(EventGoToAllianceRanking), Src: []string{StateAllianceManage}, Dst: StateAllianceRanking},

		{Name: string(EventBack), Src: []string{
			StateVIP, StateProfile, StateLeaderboard, StateSettings,
			StateChiefOrders, StateMail, StateDawnMarket,
		}, Dst: StateMainCity},
		{Name: string(EventBack), Src: []string{StateEvents}, Dst: StateMainCity},
		{Name: string(EventBack), Src: []string{StateActivityTriumph}, Dst: StateEvents},
		{Name: string(EventBack), Src: []string{
			StateAllianceHistory, StateAllianceList, StateAllianceVote, StateAllianceRanking,
			StateAllianceSettings, // 🆕 можно вернуться к alliance_manage
		}, Dst: StateAllianceManage},
		{Name: string(EventBack), Src: []string{StateAllianceManage}, Dst: StateMainCity},
		{Name: string(EventBack), Src: []string{StateExploration}, Dst: StateMainCity},
	}

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

	g.fsm = lpfsm.NewFSM(StateMainCity, transitions, callbacks)
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

func (g *GameFSM) Transition(event Event) error {
	err := g.fsm.Event(context.Background(), string(event))
	if err != nil {
		if g.logger != nil {
			g.logger.Error("FSM transition failed",
				slog.String("event", string(event)),
				slog.String("from", g.Current()),
				slog.Any("error", err),
			)
		}
		return fmt.Errorf("failed to transition on event %s from state %s: %w", event, g.Current(), err)
	}
	return nil
}

func (g *GameFSM) Current() string {
	return g.fsm.Current()
}

func (g *GameFSM) ForceTo(target string) {
	prev := g.Current()
	g.fsm.SetState(target)

	if g.logger != nil {
		g.logger.Warn("FSM forcefully moved to new state",
			slog.String("from", prev),
			slog.String("to", target),
		)
	}

	if g.callback != nil {
		before := g.getState()
		g.callback.UpdateStateFromScreenshot(target)
		after := g.getState()

		if g.logger != nil {
			utils.PrintStyledDiff(before, after)
		}
	}
}
