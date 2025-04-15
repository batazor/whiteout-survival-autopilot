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

// Researches описывает уровни исследований персонажа.
type Researches struct {
	Battle  Research `yaml:"battle"`  // Военные исследования.
	Economy Research `yaml:"economy"` // Экономические исследования.
	// Дополнительные исследования можно добавить здесь.
}

// Research описывает уровень конкретного исследования.
type Research struct {
	Level int `yaml:"level"` // Уровень исследования.
	// Дополнительные поля, если необходимо.
}
