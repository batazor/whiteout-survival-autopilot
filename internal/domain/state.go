package domain

// State содержит список аккаунтов, с которыми работает бот.
type State struct {
	Gamers Gamers `yaml:"gamers"` // Игровые персонажи
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
