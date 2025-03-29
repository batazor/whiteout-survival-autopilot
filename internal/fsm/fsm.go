package fsm

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	lpfsm "github.com/looplab/fsm"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

var fsmGraph = map[string][]string{}

func init() {
	for from, targets := range transitionPaths {
		for to := range targets {
			fsmGraph[from] = append(fsmGraph[from], to)
		}
	}
	fsmGraph[StateProfile] = append(fsmGraph[StateProfile], StateMainCity)
	fsmGraph[StateLeaderboard] = append(fsmGraph[StateLeaderboard], StateMainCity)
	fsmGraph[StateSettings] = append(fsmGraph[StateSettings], StateMainCity)
	fsmGraph[StateVIP] = append(fsmGraph[StateVIP], StateMainCity)
	fsmGraph[StateChiefOrders] = append(fsmGraph[StateChiefOrders], StateMainCity)
	fsmGraph[StateMail] = append(fsmGraph[StateMail], StateMainCity)
	fsmGraph[StateDawnMarket] = append(fsmGraph[StateDawnMarket], StateMainCity)
	fsmGraph[StateEvents] = append(fsmGraph[StateEvents], StateMainCity)
	fsmGraph[StateActivityTriumph] = append(fsmGraph[StateActivityTriumph], StateEvents)
	fsmGraph[StateAllianceSettings] = append(fsmGraph[StateAllianceSettings], StateAllianceManage)
	fsmGraph[StateAllianceHistory] = append(fsmGraph[StateAllianceHistory], StateAllianceManage)
	fsmGraph[StateAllianceList] = append(fsmGraph[StateAllianceList], StateAllianceManage)
	fsmGraph[StateAllianceVote] = append(fsmGraph[StateAllianceVote], StateAllianceManage)
	fsmGraph[StateAllianceRanking] = append(fsmGraph[StateAllianceRanking], StateAllianceManage)
	fsmGraph[StateAllianceManage] = append(fsmGraph[StateAllianceManage], StateMainCity)
	fsmGraph[StateExploration] = append(fsmGraph[StateExploration], StateMainCity)
	fsmGraph[StateExplorationBattle] = append(fsmGraph[StateExplorationBattle], StateExploration)
}

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
	StateAllianceTech      = "alliance_tech"
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
	StateExplorationBattle = "exploration_battle"
)

var AllStates = []string{
	InitialState,
	StateMainCity,
	StateActivityTriumph,
	StateAllianceManage,
	StateAllianceTech,
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
	StateExplorationBattle,
}

type TransitionStep struct {
	Action string
	Wait   time.Duration
}

var transitionPaths = map[string]map[string][]TransitionStep{
	StateMainCity: {
		StateExploration: {
			{Action: "to_exploration", Wait: 300 * time.Millisecond},
		},
		//StateEvents:         {{Action: "to_events", Wait: 300 * time.Millisecond}},
		//StateProfile:        {{Action: "to_profile", Wait: 300 * time.Millisecond}},
		//StateLeaderboard:    {{Action: "to_leaderboard", Wait: 300 * time.Millisecond}},
		//StateSettings:       {{Action: "to_settings", Wait: 300 * time.Millisecond}},
		//StateVIP:            {{Action: "to_vip", Wait: 300 * time.Millisecond}},
		//StateChiefOrders:    {{Action: "to_chief_orders", Wait: 300 * time.Millisecond}},
		//StateMail:           {{Action: "to_mail", Wait: 300 * time.Millisecond}},
		//StateDawnMarket:     {{Action: "to_dawn_market", Wait: 300 * time.Millisecond}},
		StateAllianceManage: {
			{Action: "to_alliance_manage", Wait: 300 * time.Millisecond},
		},
		//StateAllianceSettings: {
		//	{Action: "to_alliance_manage", Wait: 300 * time.Millisecond},
		//	{Action: "to_alliance_settings", Wait: 300 * time.Millisecond},
		//},
	},
	//StateEvents: {
	//	StateActivityTriumph: {{Action: "to_activity_triumph", Wait: 300 * time.Millisecond}},
	//},
	StateAllianceManage: {
		StateAllianceTech: {
			{Action: "to_alliance_tech", Wait: 300 * time.Millisecond},
		},
		//	StateAllianceHistory:  {{Action: "to_alliance_history", Wait: 300 * time.Millisecond}},
		//	StateAllianceList:     {{Action: "to_alliance_list", Wait: 300 * time.Millisecond}},
		//	StateAllianceVote:     {{Action: "to_alliance_vote", Wait: 300 * time.Millisecond}},
		//	StateAllianceRanking:  {{Action: "to_alliance_ranking", Wait: 300 * time.Millisecond}},
		//	StateAllianceSettings: {{Action: "to_alliance_settings", Wait: 300 * time.Millisecond}},
	},
	StateExploration: {
		StateExplorationBattle: {{Action: "to_exploration_battle", Wait: 300 * time.Millisecond}},
	},
}

type GameFSM struct {
	fsm           *lpfsm.FSM
	logger        *slog.Logger
	onStateChange func(state string)
	callback      StateUpdateCallback
	getState      func() *domain.State
	adb           adb.DeviceController
	lookup        *config.AreaLookup
}

func NewGameFSM(
	logger *slog.Logger,
	adb adb.DeviceController,
	lookup *config.AreaLookup,
) *GameFSM {
	g := &GameFSM{
		logger: logger,
		adb:    adb,
		lookup: lookup,
	}

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

	g.ValidateTransitionActions()

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

func (g *GameFSM) Current() string {
	return g.fsm.Current()
}

func (g *GameFSM) ForceTo(target string) {
	prev := g.Current()

	if g.adb != nil {
		steps, found := transitionPaths[prev][target]
		if !found {
			path := g.FindPath(prev, target)
			if len(path) > 1 {
				g.logger.Warn("FSM path generated dynamically", slog.Any("path", path))
				steps = g.pathToSteps(path)
				logAutoPath(path)
			} else {
				g.logger.Error("No path found to target state", slog.String("from", prev), slog.String("to", target))
			}
		}

		for _, step := range steps {
			if _, ok := g.lookup.Get(step.Action); !ok {
				panic(fmt.Sprintf("❌ Region '%s' not found in area.json", step.Action))
			}

			if err := g.adb.ClickRegion(step.Action, g.lookup); err != nil {
				g.logger.Error("Transition step failed", slog.String("action", step.Action), slog.Any("error", err))
				break
			}
			wait := step.Wait + time.Duration(rand.Intn(300)+200)*time.Millisecond
			g.logger.Debug("Waiting after action", slog.String("action", step.Action), slog.Duration("wait", wait))
			time.Sleep(wait)
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

func logAutoPath(path []string) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log dir: %v\n", err)
		return
	}

	filePath := filepath.Join(logDir, "fsm_autogenerated_paths.log")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "# Auto-generated FSM path: %s\n", time.Now().Format(time.RFC3339))
	for i := 0; i < len(path)-1; i++ {
		fmt.Fprintf(file, "// %s -> %s\n", path[i], path[i+1])
	}
	fmt.Fprintln(file)
}

func (g *GameFSM) tryTransitionVia(from string, steps []TransitionStep) error {
	g.logger.Info("Trying indirect transition", slog.String("via", from), slog.Any("steps", steps))
	for _, step := range steps {
		if _, ok := g.lookup.Get(step.Action); !ok {
			panic(fmt.Sprintf("❌ Region '%s' not found in area.json", step.Action))
		}

		if err := g.adb.ClickRegion(step.Action, g.lookup); err != nil {
			g.logger.Error("Indirect transition failed", slog.String("step", step.Action), slog.Any("error", err))
			return err
		}

		wait := step.Wait + time.Duration(rand.Intn(300)+500)*time.Millisecond
		time.Sleep(wait)
	}

	return nil
}

func (g *GameFSM) ValidateTransitionActions() {
	missing := make(map[string][]string)

	for from, targets := range transitionPaths {
		for to, steps := range targets {
			for _, step := range steps {
				if _, ok := g.lookup.Get(step.Action); !ok {
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

func (g *GameFSM) FindPath(from, to string) []string {
	visited := map[string]bool{}
	type node struct {
		state string
		path  []string
	}
	queue := []node{{from, []string{from}}}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if n.state == to {
			return n.path
		}
		visited[n.state] = true
		for _, next := range fsmGraph[n.state] {
			if !visited[next] {
				queue = append(queue, node{next, append(n.path, next)})
			}
		}
	}
	return nil
}

func (g *GameFSM) pathToSteps(path []string) []TransitionStep {
	var steps []TransitionStep
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		if s, ok := transitionPaths[from][to]; ok {
			steps = append(steps, s...)
		} else {
			fallback := from + "_back"
			steps = append(steps, TransitionStep{
				Action: fallback,
				Wait:   300 * time.Millisecond,
			})
			if g.logger != nil {
				g.logger.Warn("FSM fallback step assumed",
					slog.String("from", from),
					slog.String("to", to),
					slog.String("fallback_action", fallback),
				)
			}
		}
	}
	return steps
}
