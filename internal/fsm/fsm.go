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

func buildFSMGraph() map[string][]string {
	// 🧼 Полный сброс, чтобы избежать остатков от других FSM
	fsmGraph := map[string][]string{}

	for from, targets := range transitionPaths {
		for to := range targets {
			fsmGraph[from] = append(fsmGraph[from], to)
		}
	}

	// Дополнительные возвратные переходы
	//fsmGraph[StateProfile] = append(fsmGraph[StateProfile], StateMainCity)
	//fsmGraph[StateLeaderboard] = append(fsmGraph[StateLeaderboard], StateMainCity)
	//fsmGraph[StateSettings] = append(fsmGraph[StateSettings], StateMainCity)
	//fsmGraph[StateVIP] = append(fsmGraph[StateVIP], StateMainCity)
	//fsmGraph[StateChiefOrders] = append(fsmGraph[StateChiefOrders], StateMainCity)
	//fsmGraph[StateMail] = append(fsmGraph[StateMail], StateMainCity)
	//fsmGraph[StateDawnMarket] = append(fsmGraph[StateDawnMarket], StateMainCity)
	//fsmGraph[StateEvents] = append(fsmGraph[StateEvents], StateMainCity)
	//fsmGraph[StateActivityTriumph] = append(fsmGraph[StateActivityTriumph], StateEvents)
	//fsmGraph[StateAllianceSettings] = append(fsmGraph[StateAllianceSettings], StateAllianceManage)
	//fsmGraph[StateAllianceHistory] = append(fsmGraph[StateAllianceHistory], StateAllianceManage)
	//fsmGraph[StateAllianceList] = append(fsmGraph[StateAllianceList], StateAllianceManage)
	//fsmGraph[StateAllianceVote] = append(fsmGraph[StateAllianceVote], StateAllianceManage)
	//fsmGraph[StateAllianceRanking] = append(fsmGraph[StateAllianceRanking], StateAllianceManage)
	//fsmGraph[StateAllianceManage] = append(fsmGraph[StateAllianceManage], StateMainCity)
	//fsmGraph[StateExploration] = append(fsmGraph[StateExploration], StateMainCity)
	//fsmGraph[StateExplorationBattle] = append(fsmGraph[StateExplorationBattle], StateExploration)
	//fsmGraph[StateChiefProfile] = append(fsmGraph[StateChiefProfile], StateChiefProfileSetting)
	//fsmGraph[StateChiefProfileSetting] = append(fsmGraph[StateChiefProfileSetting], StateChiefProfileAccount)

	return fsmGraph
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

	// Смена аккаунта
	StateChiefProfile                           = "chief_profile"
	StateChiefCharacters                        = "chief_characters"
	StateChiefProfileSetting                    = "chief_profile_setting"
	StateChiefProfileAccount                    = "chief_profile_account"
	StateChiefProfileAccountChangeAccount       = "chief_profile_account_change_account"
	StateChiefProfileAccountChangeGoogle        = "chief_profile_account_change_account_google"
	StateChiefProfileAccountChangeGoogleConfirm = "chief_profile_account_change_account_google_continue"
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
	StateChiefProfile,
	StateChiefProfileSetting,
	StateChiefCharacters,
	StateChiefProfileAccount,
	StateChiefProfileAccountChangeAccount,
	StateChiefProfileAccountChangeGoogle,
	StateChiefProfileAccountChangeGoogleConfirm,
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
		StateChiefProfile: {
			{Action: "to_chief_profile", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfile: {
		StateChiefProfileSetting: {
			{Action: "to_chief_profile_setting", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfileSetting: {
		StateChiefProfileAccount: {
			{Action: "to_chief_profile_account", Wait: 300 * time.Millisecond},
		},
		StateChiefCharacters: {
			{Action: "to_chief_characters", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfileAccount: {
		StateChiefProfileAccountChangeAccount: {
			{Action: "to_change_account", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfileAccountChangeAccount: {
		StateChiefProfileAccountChangeGoogle: {
			{Action: "to_google_account", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfileAccountChangeGoogle: {
		StateChiefProfileAccountChangeGoogleConfirm: {
			{Action: "to_google_continue", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfileAccountChangeGoogleConfirm: {},
	StateChiefCharacters:                        {},
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
	fsmGraph      map[string][]string
}

func NewGame(
	logger *slog.Logger,
	adb adb.DeviceController,
	lookup *config.AreaLookup,
) *GameFSM {
	g := &GameFSM{
		logger:   logger,
		adb:      adb,
		lookup:   lookup,
		fsmGraph: buildFSMGraph(),
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

func (g *GameFSM) Current() string {
	return g.fsm.Current()
}

func (g *GameFSM) ForceTo(target string) {
	prev := g.Current()

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
				logAutoPath(path)
			} else {
				panic(fmt.Sprintf("❌ FSM: no path found from '%s' to '%s'", prev, target))
			}
		}

		for _, step := range steps {
			if _, ok := g.lookup.Get(step.Action); !ok {
				panic(fmt.Sprintf("❌ Region '%s' not found in area.json", step.Action))
			}

			if err := g.adb.ClickRegion(step.Action, g.lookup); err != nil {
				panic(fmt.Sprintf("❌ ADB click failed for action '%s': %v", step.Action, err))
			}

			wait := step.Wait + time.Duration(rand.Intn(300)+200)*time.Millisecond
			g.logger.Debug("Waiting after action", slog.String("action", step.Action), slog.Duration("wait", wait))
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

// cost возвращает стоимость ребра между состояниями.
// Если для перехода from -> to определён прямой переход, стоимость равна 1,
// иначе – предполагается fallback с стоимостью 2.
func cost(from, to string) int {
	if _, ok := transitionPaths[from][to]; ok {
		return 1
	}
	return 2
}

// FindPath ищет кратчайший путь (по суммарной стоимости ребер) от состояния from до to
// с использованием алгоритма Дейкстры.
func (g *GameFSM) FindPath(from, to string) []string {
	// Собираем все состояния: ключи графа и их соседи.
	nodes := make(map[string]bool)
	for state, neighbors := range g.fsmGraph {
		nodes[state] = true
		for _, n := range neighbors {
			nodes[n] = true
		}
	}

	const inf = int(^uint(0) >> 1)
	dist := make(map[string]int)
	prev := make(map[string]string)
	for node := range nodes {
		dist[node] = inf
	}
	dist[from] = 0

	// Множество непосещённых вершин.
	unvisited := make(map[string]bool)
	for node := range nodes {
		unvisited[node] = true
	}

	for len(unvisited) > 0 {
		var u string
		minDist := inf
		for node := range unvisited {
			if d := dist[node]; d < minDist {
				minDist = d
				u = node
			}
		}
		if u == "" {
			break // не осталось достижимых вершин
		}
		if u == to {
			break // достигли целевого состояния
		}
		delete(unvisited, u)
		for _, v := range g.fsmGraph[u] {
			if !unvisited[v] {
				continue
			}
			alt := dist[u] + cost(u, v)
			if alt < dist[v] {
				dist[v] = alt
				prev[v] = u
			}
		}
	}

	if dist[to] == inf {
		return nil // путь не найден
	}

	// Восстанавливаем путь
	var path []string
	for u := to; u != ""; u = prev[u] {
		path = append([]string{u}, path...)
		if u == from {
			break
		}
	}
	return path
}

func (g *GameFSM) pathToSteps(path []string) []TransitionStep {
	var steps []TransitionStep
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		if s, ok := transitionPaths[from][to]; ok {
			steps = append(steps, s...)
		} else {
			// Нет прямого перехода – выбрасываем ошибку
			panic(fmt.Sprintf("❌ FSM: direct transition from '%s' to '%s' not defined", from, to))
		}
	}
	return steps
}
