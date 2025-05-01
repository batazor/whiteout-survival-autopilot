package domain

import (
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/heroes"
)

type Gamers []Gamer

// Gamer описывает игрового персонажа со всеми характеристиками.
type Gamer struct {
	CurrentScreen string      `yaml:"-"`            // FSM состояние — обновляется во время игры
	ScreenState   ScreenState `yaml:"screen_state"` // Состояние экрана (например, "main", "battle", "exploration").

	ID          int           `yaml:"id"`          // Уникальный идентификатор персонажа (fid).
	Nickname    string        `yaml:"nickname"`    // Псевдоним персонажа.
	State       int           `yaml:"state"`       // Государство персонажа.
	Avatar      string        `yaml:"avatar"`      // URL аватара персонажа.
	Power       int           `yaml:"power"`       // Мощь персонажа.
	Vip_Level   int           `yaml:"vip_level"`   // VIP-уровень.
	Resources   Resources     `yaml:"resources"`   // Ресурсы персонажа.
	Gems        int           `yaml:"gems"`        // Количество гемов (премиум-валюта).
	Exploration Exploration   `yaml:"exploration"` // Исследование мира.
	Heroes      heroes.Heroes `yaml:"heroes"`      // Состояние героев.
	Messages    MessagesState `yaml:"messages"`    // Состояние сообщений.
	Alliance    Alliance      `yaml:"alliance"`    // Данные об альянсе.
	Buildings   Buildings     `yaml:"buildings"`   // Здания персонажа.
	Researches  Researches    `yaml:"researches"`  // Уровни исследований.
	Events      Events        `yaml:"events"`      // События персонажа.
	Troops      Troops        `yaml:"troops"`      // Состояние войск.
	Tech        Tech          `yaml:"tech"`        // Технологии персонажа.
}

func (g *Gamer) UpdateStateFromScreenshot(screen string) {
	g.CurrentScreen = screen
}

// Len returns the number of gamers.
func (g Gamers) Len() int {
	return len(g)
}

// Swap exchanges the gamers at indices i and j.
func (g Gamers) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// Less compares two gamers by their Nickname.
// Adjust this comparison if you want to sort by another field (e.g., ID, Power, etc.).
func (g Gamers) Less(i, j int) bool {
	return g[i].Nickname < g[j].Nickname
}
