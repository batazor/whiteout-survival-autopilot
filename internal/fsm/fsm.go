package fsm

import (
	"context"
	"fmt"
	"log/slog"

	lpfsm "github.com/looplab/fsm"
)

// --------------------------------------------------------------------
// State Definitions: Each constant represents a game screen (state)
// --------------------------------------------------------------------
const (
	// Основное состояние теперь называется main_city
	StateMainCity        = "main_city"
	StateActivityTriumph = "activity_triumph"
	StateAllianceManage  = "alliance_manage"
	StateAllianceHistory = "alliance_history"
	StateAllianceList    = "alliance_list"
	StateAllianceVote    = "alliance_vote"
	StateAllianceRanking = "alliance_ranking"
	StateEvents          = "events"
	StateProfile         = "profile"
	StateLeaderboard     = "leaderboard"
	StateSettings        = "settings"
	StateVIP             = "vip"
	StateChiefOrders     = "chief_orders"
	StateMail            = "mail"
	StateDawnMarket      = "dawn_market"
	StateExploration     = "exploration"
)

// --------------------------------------------------------------------
// Event Definitions: Each event triggers a transition between states.
// --------------------------------------------------------------------
type Event string

const (
	// Transitions from main_city to various screens
	EventGoToAllianceManage Event = "to_alliance_manage"
	EventGoToEvents         Event = "to_events"
	EventGoToProfile        Event = "to_profile"
	EventGoToLeaderboard    Event = "to_leaderboard"
	EventGoToSettings       Event = "to_settings"
	EventGoToVIP            Event = "to_vip"
	EventGoToChiefOrders    Event = "to_chief_orders"
	EventGoToMail           Event = "to_mail"
	EventGoToDawnMarket     Event = "to_dawn_market"
	EventGoToExploration    Event = "to_exploration"

	// Transition from Events screen to detailed event screen.
	EventGoToActivityTriumph Event = "to_activity_triumph"

	// Transitions within AllianceManage sub-screens.
	EventGoToAllianceHistory Event = "to_alliance_history"
	EventGoToAllianceList    Event = "to_alliance_list"
	EventGoToAllianceVote    Event = "to_alliance_vote"
	EventGoToAllianceRanking Event = "to_alliance_ranking"

	// Universal back event to return to previous or parent screen.
	EventBack Event = "back"
)

// --------------------------------------------------------------------
// GameFSM struct wraps the looplab/fsm FSM to manage game screen transitions.
// --------------------------------------------------------------------
type GameFSM struct {
	fsm    *lpfsm.FSM
	logger *slog.Logger
}

// NewGameFSM initializes and returns a new GameFSM with predefined transitions.
// The initial state is set to main_city.
func NewGameFSM(logger *slog.Logger) *GameFSM {
	// Define valid transitions using looplab/fsm.Events.
	transitions := lpfsm.Events{
		// Transitions from main_city to various screens.
		{Name: string(EventGoToAllianceManage), Src: []string{StateMainCity}, Dst: StateAllianceManage},
		{Name: string(EventGoToEvents), Src: []string{StateMainCity}, Dst: StateEvents},
		{Name: string(EventGoToProfile), Src: []string{StateMainCity}, Dst: StateProfile},
		{Name: string(EventGoToLeaderboard), Src: []string{StateMainCity}, Dst: StateLeaderboard},
		{Name: string(EventGoToSettings), Src: []string{StateMainCity}, Dst: StateSettings},
		{Name: string(EventGoToVIP), Src: []string{StateMainCity}, Dst: StateVIP},
		{Name: string(EventGoToChiefOrders), Src: []string{StateMainCity}, Dst: StateChiefOrders},
		{Name: string(EventGoToMail), Src: []string{StateMainCity}, Dst: StateMail},
		{Name: string(EventGoToDawnMarket), Src: []string{StateMainCity}, Dst: StateDawnMarket},
		{Name: string(EventGoToExploration), Src: []string{StateMainCity}, Dst: StateExploration},

		// Transition from Events to ActivityTriumph detail screen.
		{Name: string(EventGoToActivityTriumph), Src: []string{StateEvents}, Dst: StateActivityTriumph},

		// Transitions within AllianceManage sub-screens.
		{Name: string(EventGoToAllianceHistory), Src: []string{StateAllianceManage}, Dst: StateAllianceHistory},
		{Name: string(EventGoToAllianceList), Src: []string{StateAllianceManage}, Dst: StateAllianceList},
		{Name: string(EventGoToAllianceVote), Src: []string{StateAllianceManage}, Dst: StateAllianceVote},
		{Name: string(EventGoToAllianceRanking), Src: []string{StateAllianceManage}, Dst: StateAllianceRanking},

		// Back transitions:
		// Screens that return directly to main_city.
		{Name: string(EventBack), Src: []string{
			StateVIP, StateProfile, StateLeaderboard, StateSettings,
			StateChiefOrders, StateMail, StateDawnMarket,
		}, Dst: StateMainCity},
		{Name: string(EventBack), Src: []string{StateEvents}, Dst: StateMainCity},
		{Name: string(EventBack), Src: []string{StateActivityTriumph}, Dst: StateEvents},
		// Alliance sub-screens return to alliance_manage.
		{Name: string(EventBack), Src: []string{
			StateAllianceHistory, StateAllianceList, StateAllianceVote, StateAllianceRanking,
		}, Dst: StateAllianceManage},
		// Return from alliance_manage to main_city.
		{Name: string(EventBack), Src: []string{StateAllianceManage}, Dst: StateMainCity},
		// Exploration back transition.
		{Name: string(EventBack), Src: []string{StateExploration}, Dst: StateMainCity},
	}

	// Define callbacks for state transitions using the new Callback signature.
	callbacks := lpfsm.Callbacks{
		"enter_state": func(ctx context.Context, e *lpfsm.Event) {
			if logger != nil {
				logger.Info("FSM entered new state",
					slog.String("from", e.Src),
					slog.String("to", e.Dst),
					slog.String("event", e.Event),
				)
			}
		},
	}

	// Create the FSM with the initial state set to main_city.
	f := lpfsm.NewFSM(
		StateMainCity,
		transitions,
		callbacks,
	)

	return &GameFSM{fsm: f, logger: logger}
}

// Transition triggers a state transition for the given event.
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
		return fmt.Errorf("failed to transition on event %s from state %s: %w", event, g.fsm.Current(), err)
	}
	return nil
}

// Current returns the current state of the FSM.
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
}
