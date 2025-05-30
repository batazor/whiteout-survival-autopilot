package heroes

// Heroes содержит информацию о состоянии всех героев.
type Heroes struct {
	IsNotify bool `json:"isNotify"` // Признак необходимости уведомления о состоянии героев.

	List map[string]Hero
}

// Hero представляет одного героя с его характеристиками и состоянием.
type Hero struct {
	Class      string            `json:"class"`           // Infantry, Lancer, Marksman
	Generation int               `json:"generation"`      // Поколение героя
	Roles      []string          `json:"roles"`           // Роли (rally_leader, resource_gathering, и т.д.)
	Skills     HeroSkills        `json:"skills"`          // Навыки героя
	Buffs      map[string]string `json:"buffs,omitempty"` // Баффы, ключ-значение
	Notes      string            `json:"notes,omitempty"` // Примечания по использованию героя

	State State `json:"state,omitempty"` // Текущее состояние героя у пользователя
}

// HeroSkills содержит группы навыков (пока — только expedition).
type HeroSkills struct {
	Expedition map[string]HeroSkill `json:"expedition"` // one, two, three и т.д.
}

// HeroSkill представляет конкретный навык.
type HeroSkill struct {
	Name     string `json:"name"`     // Название навыка
	Priority int    `json:"priority"` // Приоритет навыка (чем выше — тем важнее)
}

// State описывает текущий статус героя у пользователя.
type State struct {
	Level         int  `json:"level"`           // Уровень прокачки
	IsAvailable   bool `json:"is_available"`    // Доступен ли герой пользователю
	IsCampTrainer bool `json:"is_camp_trainer"` // Является ли герой тренером лагеря
}
