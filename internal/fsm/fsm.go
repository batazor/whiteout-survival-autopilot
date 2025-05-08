package fsm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	lpfsm "github.com/looplab/fsm"
	"github.com/spf13/viper"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

func init() {
	mergeTransitions(transitionPaths, tundraAdventureTransitionPaths)
	mergeTransitions(transitionPaths, mainMenuTransitionPaths)
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

type StateUpdateCallback interface{}

type TransitionStep struct {
	Action  string
	Wait    time.Duration
	Trigger string // Опциональный CEL-триггер
}

var transitionPaths = map[string]map[string][]TransitionStep{
	state.StateMainCity: {
		state.StateExploration: {
			{Action: "to_exploration", Wait: 300 * time.Millisecond},
		},
		//StateEvents:         {{Action: "to_events", Wait: 300 * time.Millisecond}},
		//StateProfile:        {{Action: "to_profile", Wait: 300 * time.Millisecond}},
		//StateLeaderboard:    {{Action: "to_leaderboard", Wait: 300 * time.Millisecond}},
		//StateSettings:       {{Action: "to_settings", Wait: 300 * time.Millisecond}},
		//StateVIP:            {{Action: "to_vip", Wait: 300 * time.Millisecond}},
		//StateChiefOrders:    {{Action: "to_chief_orders", Wait: 300 * time.Millisecond}},
		//StateDawnMarket:     {{Action: "to_dawn_market", Wait: 300 * time.Millisecond}},
		state.StateAllianceManage: {
			{Action: "to_alliance_manage", Wait: 300 * time.Millisecond},
		},
		//StateAllianceSettings: {
		//	{Action: "to_alliance_manage", Wait: 300 * time.Millisecond},
		//	{Action: "to_alliance_settings", Wait: 300 * time.Millisecond},
		//},
		state.StateChiefProfile: {
			{Action: "to_chief_profile", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
		state.StatePets: {
			{Action: "to_pets", Wait: 300 * time.Millisecond},
		},
		state.StateWorld: {
			{Action: "to_world", Wait: 300 * time.Millisecond},
		},
		state.StateMail: {
			{Action: "to_mail", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventure: {
			{
				Action:  "events.tundraAdventure.state.isExist",
				Wait:    300 * time.Millisecond,
				Trigger: "events.tundraAdventure.state.isExist",
			},
		},
		state.StateVIP: {
			{Action: "to_vip", Wait: 300 * time.Millisecond},
		},
		state.StateChiefOrders: {
			{Action: "to_chief_orders", Wait: 300 * time.Millisecond},
		},
		state.StateDailyMissions: {
			{Action: "to_daily_missions", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMainMenuWilderness: {
		state.StateMainMenuCity: {
			{Action: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfile: {
		state.StateChiefProfileSetting: {
			{Action: "to_chief_profile_setting", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Action: "chief_profile_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileSetting: {
		state.StateChiefProfileAccount: {
			{Action: "to_chief_profile_account", Wait: 300 * time.Millisecond},
		},
		state.StateChiefCharacters: {
			{Action: "to_chief_characters", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccount: {
		state.StateChiefProfileAccountChangeAccount: {
			{Action: "to_change_account", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccountChangeAccount: {
		state.StateChiefProfileAccountChangeGoogle: {
			{Action: "to_google_account", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccountChangeGoogle: {
		state.StateChiefProfileAccountChangeGoogleConfirm: {
			{Action: "to_google_continue", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccountChangeGoogleConfirm: {},
	state.StateChiefCharacters:                        {},
	//StateEvents: {
	//	state.StateActivityTriumph: {{Action: "to_activity_triumph", Wait: 300 * time.Millisecond}},
	//},
	state.StateAllianceManage: {
		state.StateAllianceTech: {
			{Action: "to_alliance_tech", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceRanking:  {{Action: "to_alliance_power_rankings", Wait: 300 * time.Millisecond}},
		state.StateAllianceSettings: {{Action: "to_alliance_settings", Wait: 300 * time.Millisecond}},
		state.StateMainCity: {
			{Action: "to_alliance_back", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWar: {
			{Action: "to_alliance_war", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChests: {
			{Action: "to_alliance_chests", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceChests: {
		state.StateAllianceChestLoot: {
			{Action: "to_alliance_chest_loot", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChestGift: {
			{Action: "to_alliance_chest_gift", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Action: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceChestLoot: {
		state.StateAllianceManage: {
			{Action: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChestGift: {
			{Action: "to_alliance_chest_gift", Wait: 100 * time.Millisecond},
		},
	},
	state.StateAllianceChestGift: {
		state.StateAllianceManage: {
			{Action: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChestLoot: {
			{Action: "to_alliance_chest_loot", Wait: 100 * time.Millisecond},
		},
	},
	state.StateAllianceTech: {
		state.StateAllianceManage: {
			{Action: "from_tech_to_alliance", Wait: 300 * time.Millisecond},
		},
	},
	state.StateExploration: {
		state.StateExplorationBattle: {{Action: "to_exploration_battle", Wait: 300 * time.Millisecond}},
		state.StateMainCity:          {{Action: "exploration_back", Wait: 300 * time.Millisecond}},
	},
	state.StateExplorationBattle: {
		state.StateExploration: {{Action: "to_exploration_battle_back", Wait: 300 * time.Millisecond}},
	},
	state.StateTopUpCenter: {
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceWar: {
		state.StateAllianceWarRally: {
			{Action: "to_alliance_war_rally", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWarSolo: {
			{Action: "to_alliance_war_solo", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWarEvents: {
			{Action: "to_alliance_war_events", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Action: "from_war_to_alliance_manage", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceWarRally: {
		state.StateAllianceWarRallyAutoJoin: {
			{Action: "to_alliance_war_auto_join", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWar: {},
		state.StateAllianceWarSolo: {
			{Action: "to_alliance_war_solo", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWarEvents: {
			{Action: "to_alliance_war_events", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Action: "from_war_to_alliance_manage", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceWarRallyAutoJoin: {
		state.StateAllianceWar: {
			{Action: "alliance_war_auto_join_close", Wait: 300 * time.Millisecond},
		},
	},
	state.StateWorld: {
		state.StateMainCity: {
			{Action: "to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateWorldSearch: {
			{Action: "to_search_resources", Wait: 300 * time.Millisecond},
		},
		state.StateWorldGlobalMap: {
			{Action: "to_global_map", Wait: 300 * time.Millisecond},
		},
		state.StateMail: {
			{Action: "to_mail", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMail: {
		state.StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailWars: {
		state.StateMail: {},
		state.StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailAlliance: {
		state.StateMail: {},
		state.StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailSystem: {
		state.StateMail: {},
		state.StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailReports: {
		state.StateMail: {},
		state.StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailStarred: {
		state.StateMail: {},
		state.StateMainCity: {
			{Action: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Action: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Action: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Action: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Action: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Action: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateVIP: {
		state.StateMainCity: {
			{Action: "from_vip_to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateVIPAdd: {
			{Action: "to_vip_add", Wait: 300 * time.Millisecond},
		},
	},
	state.StateVIPAdd: {
		state.StateVIP: {
			{Action: "from_vip_add_to_vip", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefOrders: {
		state.StateMainCity: {
			{Action: "from_chief_orders_to_main_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateDailyMissions: {
		state.StateMainCity: {
			{
				Action: "from_daily_missions_to_main_city",
				Wait:   300 * time.Millisecond,
			},
		},
		state.StateGrowthMissions: {
			{Action: "to_growth_missions", Wait: 300 * time.Millisecond},
		},
	},
	state.StateGrowthMissions: {
		state.StateMainCity: {
			{Action: "from_growth_missions_to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateDailyMissions: {
			{Action: "from_growth_missions_to_daily_missions", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpack: {
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackResources: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackSpeedups: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackBonus: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackGear: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackOther: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChat: {
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChatAlliance: {
		state.StateChat: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChatWorld: {
		state.StateChat: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChatPersonal: {
		state.StateChat: {},
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateHeroes: {
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateEvents: {
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateDeals: {
		state.StateMainCity: {
			{Action: "page_back", Wait: 300 * time.Millisecond},
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
	rulesCheckState  config.ScreenAnalyzeRules

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
	viper.AutomaticEnv()

	viper.SetDefault("PATH_TO_FSM_STATE_RULES", "references/fsmState.yaml")
	pathToFSMStateRules := viper.GetString("PATH_TO_FSM_STATE_RULES")

	// ─── Инициализация правил проверки состояния ───────────────────────
	rulesCheckState, err := config.LoadAnalyzeRules(pathToFSMStateRules)
	if err != nil {
		logger.Error("Failed to load analyze rules", slog.Any("error", err))

		panic("Failed to load analyze rules")
	}

	// Начинаем с главного экрана
	if gamerState != nil {
		gamerState.ScreenState.CurrentState = state.StateMainCity
	}

	g := &GameFSM{
		logger:           logger,
		adb:              adb,
		lookup:           lookup,
		fsmGraph:         buildFSMGraph(),
		triggerEvaluator: triggerEvaluator,
		gamerState:       gamerState,
		rulesCheckState:  rulesCheckState,
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

	g.fsm = lpfsm.NewFSM(state.StateMainCity, transitions, callbacks)
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
