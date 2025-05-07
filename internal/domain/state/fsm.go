package state

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

	// Губернатор
	StateChiefOrders = "chief_orders"

	// Ежедневные задания
	StateDailyMissions = "daily_missions"
	// Миссии роста
	StateGrowthMissions = "growth_missions"
)

const (
	// Tundra Adventure
	StateTundraAdventure               = "tundra_adventure"
	StateTundraAdventureMain           = "tundra_adventure_main"
	StateTundraAdventureDrill          = "tundra_adventure_drill"
	StateTundraAdventurerDrill         = "tundra_adventurer_drill"
	StateTundraAdventurerDailyMissions = "tundra_adventurer_daily_missions"
	StateTundraAdventureOdessey        = "tundra_adventure_odessey"
	StateTundraAdventureCaravan        = "tundra_adventure_caravan"
)

const (
	StateInfantryCityView = "infantry_city_view"
	StateLancerCityView   = "lancer_city_view"
	StateMarksmanCityView = "marksman_city_view"
)

const (
	// Главное меню
	StateMainMenuCity         = "main_menu_city"
	StateMainMenuWilderness   = "main_menu_wilderness"
	StateMainMenuBuilding1    = "main_menu_building_1"
	StateMainMenuBuilding2    = "main_menu_building_2"
	StateMainMenuTechResearch = "main_menu_tech_research"
)
