package fsm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	lpfsm "github.com/looplab/fsm"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func init() {
	mergeTransitions(transitionPaths, tundraAdventureTransitionPaths)
	mergeTransitions(transitionPaths, mainMenuTransitionPaths)
	mergeTransitions(transitionPaths, troopsTransitionPaths)
	mergeTransitions(transitionPaths, troopsTransitionPaths)
}

func buildFSMGraph() map[string][]string {
	// 🧼 Полный сброс, чтобы избежать остатков от других FSM
	fsmGraph := map[string][]string{}

	for from, targets := range transitionPaths {
		for to := range targets {
			fsmGraph[from] = append(fsmGraph[from], to)
		}
	}

	return fsmGraph
}

type StateUpdateCallback interface {
	UpdateStateFromScreenshot(screen string)
}

// --------------------------------------------------------------------
// State Definitions: Each constant represents a game screen (state)
// --------------------------------------------------------------------
const (
	InitialState         = "initial"
	StateMainCity        = "main_city"
	StateActivityTriumph = "activity_triumph"
	StateEvents          = "events"
	StateProfile         = "profile"
	StateLeaderboard     = "leaderboard"
	StateSettings        = "settings"
	StateChiefOrders     = "chief_orders"
	StateDawnMarket      = "dawn_market"

	// Питомцы
	StatePets = "pets"

	// Исследование
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

	// Альянс
	StateAllianceManage      = "alliance_manage"
	StateAllianceTech        = "alliance_tech"
	StateAllianceSettings    = "alliance_settings"
	StateAllianceRanking     = "alliance_ranking"
	StateAllianceWar         = "alliance_war"
	StateAllianceWarAutoJoin = "alliance_war_auto_join"

	// Альянс - сундуки
	StateAllianceChests    = "alliance_chests"
	StateAllianceChestLoot = "alliance_chest_loot"
	StateAllianceChestGift = "alliance_chest_gift"

	// Глобальная карта
	StateWorld          = "world"
	StateWorldSearch    = "world_search_resources"
	StateWorldGlobalMap = "world_global_map"

	// Сообщения
	StateMail         = "mail"
	StateMailWars     = "mail_wars"
	StateMailAlliance = "mail_alliance"
	StateMailSystem   = "mail_system"
	StateMailReports  = "mail_reports"
	StateMailStarred  = "mail_starred"

	// VIP
	StateVIP    = "vip"
	StateVIPAdd = "vip_add"
)

type TransitionStep struct {
	Action  string
	Wait    time.Duration
	Trigger string // Опциональный CEL-триггер
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
		StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
		StatePets: {
			{Action: "to_pets", Wait: 300 * time.Millisecond},
		},
		StateWorld: {
			{Action: "to_world", Wait: 300 * time.Millisecond},
		},
		StateMail: {
			{Action: "to_mail", Wait: 300 * time.Millisecond},
		},
		StateTundraAdventure: {
			{
				Action:  "events.tundraAdventure.state.isExist",
				Wait:    300 * time.Millisecond,
				Trigger: "events.tundraAdventure.state.isExist",
			},
		},
		StateVIP: {
			{Action: "to_vip", Wait: 300 * time.Millisecond},
		},
	},
	StateMainMenuWilderness: {
		StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	StateChiefProfile: {
		StateChiefProfileSetting: {
			{Action: "to_chief_profile_setting", Wait: 300 * time.Millisecond},
		},
		StateMainCity: {
			{Action: "chief_profile_back", Wait: 300 * time.Millisecond},
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
		StateAllianceRanking:  {{Action: "to_alliance_power_rankings", Wait: 300 * time.Millisecond}},
		StateAllianceSettings: {{Action: "to_alliance_settings", Wait: 300 * time.Millisecond}},
		StateMainCity: {
			{Action: "to_alliance_back", Wait: 300 * time.Millisecond},
		},
		StateAllianceWar: {
			{Action: "to_alliance_war", Wait: 300 * time.Millisecond},
		},
		StateAllianceChests: {
			{Action: "to_alliance_chests", Wait: 300 * time.Millisecond},
		},
	},
	StateAllianceChests: {
		StateAllianceChestLoot: {
			{Action: "to_alliance_chest_loot", Wait: 300 * time.Millisecond},
		},
		StateAllianceChestGift: {
			{Action: "to_alliance_chest_gift", Wait: 300 * time.Millisecond},
		},
		StateAllianceManage: {
			{Action: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
	},
	StateAllianceChestLoot: {
		StateAllianceManage: {
			{Action: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
		StateAllianceChestGift: {
			{Action: "to_alliance_chest_gift", Wait: 100 * time.Millisecond},
		},
	},
	StateAllianceChestGift: {
		StateAllianceManage: {
			{Action: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
		StateAllianceChestLoot: {
			{Action: "to_alliance_chest_loot", Wait: 100 * time.Millisecond},
		},
	},
	StateAllianceTech: {
		StateAllianceManage: {
			{Action: "from_tech_to_alliance", Wait: 300 * time.Millisecond},
		},
	},
	StateExploration: {
		StateExplorationBattle: {{Action: "to_exploration_battle", Wait: 300 * time.Millisecond}},
		StateMainCity:          {{Action: "exploration_back", Wait: 300 * time.Millisecond}},
	},
	StateExplorationBattle: {
		StateExploration: {{Action: "to_exploration_battle_back", Wait: 300 * time.Millisecond}},
	},
	StateAllianceWar: {
		StateAllianceWarAutoJoin: {
			{Action: "to_alliance_war_auto_join", Wait: 300 * time.Millisecond},
		},
		StateAllianceManage: {
			{Action: "from_war_to_alliance_manage", Wait: 300 * time.Millisecond},
		},
	},
	StateAllianceWarAutoJoin: {
		StateAllianceWar: {
			{Action: "alliance_war_auto_join_close", Wait: 300 * time.Millisecond},
		},
	},
	StateWorld: {
		StateMainCity: {
			{Action: "to_main_city", Wait: 300 * time.Millisecond},
		},
		StateWorldSearch: {
			{Action: "to_search_resources", Wait: 300 * time.Millisecond},
		},
		StateWorldGlobalMap: {
			{Action: "to_global_map", Wait: 300 * time.Millisecond},
		},
		StateMail: {
			{Action: "to_mail", Wait: 300 * time.Millisecond},
		},
	},
	StateMail: {
		StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	StateMailWars: {
		StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	StateMailAlliance: {
		StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	StateMailSystem: {
		StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	StateMailReports: {
		StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	StateMailStarred: {
		StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	StateVIP: {
		StateMainCity: {
			{Action: "from_vip_to_main_city", Wait: 300 * time.Millisecond},
		},
		StateVIPAdd: {
			{Action: "to_vip_add", Wait: 300 * time.Millisecond},
		},
	},
	StateVIPAdd: {
		StateVIP: {
			{Action: "from_vip_add_to_vip", Wait: 300 * time.Millisecond},
		},
	},
}

type GameFSM struct {
	fsm              *lpfsm.FSM
	analyzer         *analyzer.Analyzer
	logger           *slog.Logger
	onStateChange    func(state string)
	callback         StateUpdateCallback
	gamerState       *domain.Gamer
	adb              adb.DeviceController
	lookup           *config.AreaLookup
	fsmGraph         map[string][]string
	triggerEvaluator config.TriggerEvaluator

	// previousState хранит предыдущее состояние FSM
	previousState string
}

func NewGame(
	logger *slog.Logger,
	adb adb.DeviceController,
	lookup *config.AreaLookup,
	triggerEvaluator config.TriggerEvaluator,
	gamerState *domain.Gamer,
) *GameFSM {
	g := &GameFSM{
		logger:           logger,
		adb:              adb,
		lookup:           lookup,
		fsmGraph:         buildFSMGraph(),
		triggerEvaluator: triggerEvaluator,
		gamerState:       gamerState,
		analyzer:         analyzer.NewAnalyzer(lookup, logger),
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

	if gs, ok := cb.(*domain.Gamer); ok {
		g.gamerState = gs
	}
}

func (g *GameFSM) SetOnStateChange(f func(state string)) {
	g.onStateChange = f
}

func (g *GameFSM) Current() string {
	return g.fsm.Current()
}

func (g *GameFSM) logAutoPath(path []string) {
	if len(path) < 2 {
		return
	}

	g.logger.Info("📍 Auto-generated FSM path", slog.String("timestamp", time.Now().Format(time.RFC3339)))

	for i := 0; i < len(path)-1; i++ {
		from, to := path[i], path[i+1]
		g.logger.Info("→ FSM step", slog.String("from", from), slog.String("to", to))
	}
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

func mergeTransitions(dst, src map[string]map[string][]TransitionStep) {
	for from, targets := range src {
		if _, ok := dst[from]; !ok {
			dst[from] = make(map[string][]TransitionStep)
		}
		for to, steps := range targets {
			dst[from][to] = steps
		}
	}
}
