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
	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
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

// TransitionStep описывает один шаг перехода FSM между экранами.
type TransitionStep struct {
	Click   string        // Имя региона для клика (оставьте пустым, если не нужен)
	Swipe   *Swipe        // Описание свайпа (nil, если свайп не требуется)
	Wait    time.Duration // Время ожидания после шага
	Trigger string        // Опциональный CEL-триггер для условия выполнения шага
}

// Swipe позволяет указывать свайпы двумя способами:
// 1. Абсолютные координаты (X1, Y1, X2, Y2)
// 2. Декларативно: через Direction (направление) и Delta (смещение в пикселях).
// Если Direction заполнен, координаты вычисляются автоматически из центра экрана.
//
// Пример декларативного свайпа: &Swipe{Direction: "left", Delta: 500, Duration: 350 * time.Millisecond}
//
// Direction: "left", "right", "up", "down" (игнорируется, если используются X1..Y2)
// Delta: на сколько пикселей сдвигать (для Direction)
type Swipe struct {
	X1, Y1, X2, Y2 int    // Абсолютные координаты свайпа (если Direction пуст)
	Direction      string // "left", "right", "up", "down". Если не пусто, используются Delta и Duration
	Delta          int    // Смещение в пикселях для Direction
}

var transitionPaths = map[string]map[string][]TransitionStep{
	state.StateMainCity: {
		state.StateExploration: {
			{Click: "to_exploration", Wait: 300 * time.Millisecond},
		},
		//StateEvents:         {{Click: "to_events", Wait: 300 * time.Millisecond}},
		//StateProfile:        {{Click: "to_profile", Wait: 300 * time.Millisecond}},
		//StateLeaderboard:    {{Click: "to_leaderboard", Wait: 300 * time.Millisecond}},
		//StateSettings:       {{Click: "to_settings", Wait: 300 * time.Millisecond}},
		//StateVIP:            {{Click: "to_vip", Wait: 300 * time.Millisecond}},
		//StateChiefOrders:    {{Click: "to_chief_orders", Wait: 300 * time.Millisecond}},
		//StateDawnMarket:     {{Click: "to_dawn_market", Wait: 300 * time.Millisecond}},
		state.StateAllianceManage: {
			{Click: "to_alliance_manage", Wait: 300 * time.Millisecond},
		},
		//StateAllianceSettings: {
		//	{Click: "to_alliance_manage", Wait: 300 * time.Millisecond},
		//	{Click: "to_alliance_settings", Wait: 300 * time.Millisecond},
		//},
		state.StateChiefProfile: {
			{Click: "to_chief_profile", Wait: 300 * time.Millisecond},
		},
		state.StateMainMenuCity: {
			{Click: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
		state.StatePets: {
			{Click: "to_pets", Wait: 300 * time.Millisecond},
		},
		state.StateWorld: {
			{Click: "to_world", Wait: 300 * time.Millisecond},
		},
		state.StateMail: {
			{Click: "to_mail", Wait: 300 * time.Millisecond},
		},
		state.StateTundraAdventure: {
			{
				Click:   "events.tundraAdventure.state.isExist",
				Wait:    300 * time.Millisecond,
				Trigger: "events.tundraAdventure.state.isExist",
			},
		},
		state.StateVIP: {
			{Click: "to_vip", Wait: 300 * time.Millisecond},
		},
		state.StateChiefOrders: {
			{Click: "to_chief_orders", Wait: 300 * time.Millisecond},
		},
		state.StateDailyMissions: {
			{Click: "to_daily_missions", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMainMenuWilderness: {
		state.StateMainMenuCity: {
			{Click: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfile: {
		state.StateChiefProfileSetting: {
			{Click: "to_chief_profile_setting", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "chief_profile_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileSetting: {
		state.StateChiefProfileAccount: {
			{Click: "to_chief_profile_account", Wait: 300 * time.Millisecond},
		},
		state.StateChiefCharacters: {
			{Click: "to_chief_characters", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccount: {
		state.StateChiefProfileAccountChangeAccount: {
			{Click: "to_change_account", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccountChangeAccount: {
		state.StateChiefProfileAccountChangeGoogle: {
			{Click: "to_google_account", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccountChangeGoogle: {
		state.StateChiefProfileAccountChangeGoogleConfirm: {
			{Click: "to_google_continue", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefProfileAccountChangeGoogleConfirm: {},
	state.StateChiefCharacters:                        {},
	//StateEvents: {
	//	state.StateActivityTriumph: {{Click: "to_activity_triumph", Wait: 300 * time.Millisecond}},
	//},
	state.StateAllianceManage: {
		state.StateAllianceTech: {
			{Click: "to_alliance_tech", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceRanking:  {{Click: "to_alliance_power_rankings", Wait: 300 * time.Millisecond}},
		state.StateAllianceSettings: {{Click: "to_alliance_settings", Wait: 300 * time.Millisecond}},
		state.StateMainCity: {
			{Click: "to_alliance_back", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWar: {
			{Click: "to_alliance_war", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChests: {
			{Click: "to_alliance_chests", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceTerritory: {
		state.StateAllianceManage: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceChests: {
		state.StateAllianceChestLoot: {
			{Click: "to_alliance_chest_loot", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChestGift: {
			{Click: "to_alliance_chest_gift", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Click: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceChestLoot: {
		state.StateAllianceManage: {
			{Click: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChestGift: {
			{Click: "to_alliance_chest_gift", Wait: 100 * time.Millisecond},
		},
	},
	state.StateAllianceChestGift: {
		state.StateAllianceManage: {
			{Click: "to_alliance_manager_from_alliance_chest", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceChestLoot: {
			{Click: "to_alliance_chest_loot", Wait: 100 * time.Millisecond},
		},
	},
	state.StateAllianceTech: {
		state.StateAllianceManage: {
			{Click: "from_tech_to_alliance", Wait: 300 * time.Millisecond},
		},
	},
	state.StateExploration: {
		state.StateExplorationBattle: {{Click: "to_exploration_battle", Wait: 300 * time.Millisecond}},
		state.StateMainCity:          {{Click: "exploration_back", Wait: 300 * time.Millisecond}},
	},
	state.StateExplorationBattle: {
		state.StateExploration: {{Click: "to_exploration_battle_back", Wait: 300 * time.Millisecond}},
	},
	state.StateTopUpCenter: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceWar: {
		state.StateAllianceWarRally: {
			{Click: "to_alliance_war_rally", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWarSolo: {
			{Click: "to_alliance_war_solo", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWarEvents: {
			{Click: "to_alliance_war_events", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Click: "from_war_to_alliance_manage", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceWarRally: {
		state.StateAllianceWarRallyAutoJoin: {
			{Click: "to_alliance_war_auto_join", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWar: {},
		state.StateAllianceWarSolo: {
			{Click: "to_alliance_war_solo", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceWarEvents: {
			{Click: "to_alliance_war_events", Wait: 300 * time.Millisecond},
		},
		state.StateAllianceManage: {
			{Click: "from_war_to_alliance_manage", Wait: 300 * time.Millisecond},
		},
	},
	state.StateAllianceWarRallyAutoJoin: {
		state.StateAllianceWar: {
			{Click: "alliance_war_auto_join_close", Wait: 300 * time.Millisecond},
		},
	},
	state.StateWorld: {
		state.StateMainCity: {
			{Click: "to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateWorldSearch: {
			{Click: "to_search_resources", Wait: 300 * time.Millisecond},
		},
		state.StateWorldGlobalMap: {
			{Click: "to_global_map", Wait: 300 * time.Millisecond},
		},
		state.StateMail: {
			{Click: "to_mail", Wait: 300 * time.Millisecond},
		},
		state.StateHealInjured: {
			{
				Click:   "healInjured.state.isAvailable",
				Wait:    300 * time.Millisecond,
				Trigger: "healInjured.state.isAvailable",
			},
		},
	},
	state.StateMail: {
		state.StateMainCity: {
			{Click: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Click: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Click: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Click: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Click: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Click: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailWars: {
		state.StateMail: {},
		state.StateMainCity: {
			{Click: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Click: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Click: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Click: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Click: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Click: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailAlliance: {
		state.StateMail: {},
		state.StateMainCity: {
			{Click: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Click: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Click: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Click: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Click: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Click: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailSystem: {
		state.StateMail: {},
		state.StateMainCity: {
			{Click: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Click: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Click: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Click: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Click: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Click: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailReports: {
		state.StateMail: {},
		state.StateMainCity: {
			{Click: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Click: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Click: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Click: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Click: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Click: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateMailStarred: {
		state.StateMail: {},
		state.StateMainCity: {
			{Click: "mail_close", Wait: 300 * time.Millisecond},
		},
		state.StateMailWars: {
			{Click: "to_mail_wars", Wait: 300 * time.Millisecond},
		},
		state.StateMailAlliance: {
			{Click: "to_mail_alliance", Wait: 300 * time.Millisecond},
		},
		state.StateMailSystem: {
			{Click: "to_mail_system", Wait: 300 * time.Millisecond},
		},
		state.StateMailReports: {
			{Click: "to_mail_reports", Wait: 300 * time.Millisecond},
		},
		state.StateMailStarred: {
			{Click: "to_mail_starred", Wait: 300 * time.Millisecond},
		},
	},
	state.StateVIP: {
		state.StateMainCity: {
			{Click: "from_vip_to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateVIPAdd: {
			{Click: "to_vip_add", Wait: 300 * time.Millisecond},
		},
	},
	state.StateVIPAdd: {
		state.StateVIP: {
			{Click: "from_vip_add_to_vip", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChiefOrders: {
		state.StateMainCity: {
			{Click: "from_chief_orders_to_main_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateDailyMissions: {
		state.StateMainCity: {
			{
				Click: "from_daily_missions_to_main_city",
				Wait:  300 * time.Millisecond,
			},
		},
		state.StateGrowthMissions: {
			{Click: "to_growth_missions", Wait: 300 * time.Millisecond},
		},
	},
	state.StateGrowthMissions: {
		state.StateMainCity: {
			{Click: "from_growth_missions_to_main_city", Wait: 300 * time.Millisecond},
		},
		state.StateDailyMissions: {
			{Click: "from_growth_missions_to_daily_missions", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpack: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackResources: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackSpeedups: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackBonus: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackGear: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateBackpackOther: {
		state.StateBackpack: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChat: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChatAlliance: {
		state.StateChat: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChatWorld: {
		state.StateChat: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateChatPersonal: {
		state.StateChat: {},
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateHeroes: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateNatalia: {
		state.StateHeroes: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateEvents: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateDeals: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateIntel: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateArenaCityView: {
		state.StateMainCity: {},
		state.StateMainMenuCity: {
			{Click: "to_main_menu_city", Wait: 300 * time.Millisecond},
		},
		state.StateArenaMain: {
			{Click: "to_arena_main", Wait: 2 * time.Second},
		},
	},
	state.StateArenaMain: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
		state.StateArenaDefensiveSquadLineup: {
			{Click: "to_arena_defensive_squad_lineup", Wait: 300 * time.Millisecond},
		},
		state.StateArenaChallengeList: {
			{Click: "to_arena_challenge_list", Wait: 300 * time.Millisecond},
		},
	},
	state.StateArenaDefensiveSquadLineup: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateHealInjured: {
		state.StateWorld: {
			{Click: "from_heal_injured_to_world", Wait: 300 * time.Millisecond},
		},
		state.StateMainCity: {
			{Click: "to_main_city", Wait: 300 * time.Millisecond},
		},
	},
	state.StateActivityTriumph: {
		state.StateAllianceManage: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateLabyrinth: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateCaveOfMonsters: {
		state.StateLabyrinth: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
		},
	},
	state.StateEnlistmentOffice: {
		state.StateMainCity: {
			{Click: "page_back", Wait: 300 * time.Millisecond},
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
	OCRClient        *ocrclient.Client

	// previousState хранит предыдущее состояние FSM
	previousState string
}

func NewGame(
	logger *slog.Logger,
	adb adb.DeviceController,
	lookup *config.AreaLookup,
	triggerEvaluator config.TriggerEvaluator,
	gamerState *domain.Gamer,
	OCRClient *ocrclient.Client,
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
		analyzer:         analyzer.NewAnalyzer(lookup, logger, OCRClient),
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
				switch {
				case step.Click == "" && step.Swipe == nil:
					missing[from] = append(missing[from], fmt.Sprintf("%s → %s: missing Click and Swipe", from, to))
				case step.Click != "" && step.Swipe == nil:
					if _, ok := g.lookup.Get(step.Click); !ok {
						missing[from] = append(missing[from], fmt.Sprintf("%s → %s: '%s'", from, to, step.Click))
					}
				}
			}
		}
	}

	if len(missing) == 0 {
		return
	}

	errMsg := "❌ Missing required region definitions in area.json:\n"
	for _, issues := range missing {
		for _, entry := range issues {
			errMsg += " - " + entry + "\n"
		}
	}
	panic(errMsg)
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
