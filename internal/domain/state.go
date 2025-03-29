package domain

// State содержит список аккаунтов, с которыми работает бот.
type State struct {
	Accounts []Account `yaml:"accounts"`
}

// Account описывает аккаунт с электронной почтой и списком игровых персонажей (gamers).
type Account struct {
	Email      string  `yaml:"email"`      // Электронная почта аккаунта.
	Characters []Gamer `yaml:"characters"` // Игровые персонажи, связанные с аккаунтом.
}

// Gamer описывает игрового персонажа со всеми характеристиками.
type Gamer struct {
	ID          int           `yaml:"id"`          // Уникальный идентификатор персонажа (fid).
	Nickname    string        `yaml:"nickname"`    // Псевдоним персонажа.
	State       int           `yaml:"state"`       // Государство персонажа.
	Avatar      string        `yaml:"avatar"`      // URL аватара персонажа.
	Power       int           `yaml:"power"`       // Мощь персонажа.
	Vip_Level   int           `yaml:"vip_level"`   // VIP-уровень.
	Resources   Resources     `yaml:"resources"`   // Ресурсы персонажа.
	Gems        int           `yaml:"gems"`        // Количество гемов (премиум-валюта).
	Exploration Exploration   `yaml:"exploration"` // Исследование мира.
	Heroes      HeroesState   `yaml:"heroes"`      // Состояние героев.
	Messages    MessagesState `yaml:"messages"`    // Состояние сообщений.
	Alliance    Alliance      `yaml:"alliance"`    // Данные об альянсе.
	Buildings   Buildings     `yaml:"buildings"`   // Здания персонажа.
	Researchs   Researchs     `yaml:"researchs"`   // Уровни исследований.
}

// Exploration описывает уровень исследования мира.
type Exploration struct {
	Level int              `yaml:"level"`
	State ExplorationState `yaml:"state"`
}

// ExplorationState описывает состояние исследования мира.
type ExplorationState struct {
	IsClaimActive bool   `yaml:"isClaimActive"` // Признак доступности кнопки "Забрать".
	MyPower       int    `yaml:"myPower"`
	EnemyPower    int    `yaml:"enemyPower"`
	BattleStatus  string `yaml:"battleStatus"`
}

// Resources описывает ресурсы, которыми владеет персонаж.
type Resources struct {
	Wood int `yaml:"wood"` // Дерево.
	Food int `yaml:"food"` // Продовольствие.
	Iron int `yaml:"iron"` // Железо.
	Meat int `yaml:"meat"` // Мясо.
}

// HeroesState содержит информацию о состоянии героев.
type HeroesState struct {
	State HeroesStatus `yaml:"state"`
}

// HeroesStatus описывает статус героев.
type HeroesStatus struct {
	IsHeroes bool `yaml:"isHeroes"` // Признак активности героев.
}

// MessagesState содержит информацию о сообщениях персонажа.
type MessagesState struct {
	State MessageStatus `yaml:"state"`
}

// MessageStatus описывает состояние сообщений.
type MessageStatus struct {
	IsNewMessage bool `yaml:"isNewMessage"` // Признак наличия новых сообщений.
	IsNewReports bool `yaml:"isNewReports"` // Признак наличия новых отчётов.
}

// Buildings представляет собой набор зданий персонажа.
type Buildings struct {
	Furnace Building `yaml:"furnace"` // Печь.
}

// Building описывает отдельное здание.
type Building struct {
	Level int `yaml:"level"` // Уровень здания.
	Power int `yaml:"power"` // Мощь здания.
	// Можно добавить дополнительные поля, например, время постройки, затраты ресурсов и т.д.
}

// Researchs описывает уровни исследований персонажа.
type Researchs struct {
	Battle  Research `yaml:"battle"`  // Военные исследования.
	Economy Research `yaml:"economy"` // Экономические исследования.
	// Дополнительные исследования можно добавить здесь.
}

// Research описывает уровень конкретного исследования.
type Research struct {
	Level int `yaml:"level"` // Уровень исследования.
	// Дополнительные поля, если необходимо.
}
