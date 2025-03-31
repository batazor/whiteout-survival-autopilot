package domain

// Gamer описывает игрового персонажа со всеми характеристиками.
type Gamer struct {
	CurrentScreen string `yaml:"-"` // FSM состояние — обновляется во время игры

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

func (g *Gamer) UpdateStateFromScreenshot(screen string) {
	g.CurrentScreen = screen
}
