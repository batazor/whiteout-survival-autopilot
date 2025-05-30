package domain

type HealInjured struct {
	State HealInjuredState `yaml:"state"` // Состояние лечения раненых
}

type HealInjuredState struct {
	IsAvailable bool   `yaml:"isAvailable"` // Признак доступности лечения раненых (по иконке)
	IsNext      string `yaml:"isNext"`      // Признак доступности лечения раненых (по тексту)

	IsReplenishAll bool   `yaml:"isReplenishAll"` // Признак доступности лечения всех раненых
	StatusHeal     string `yaml:"statusHeal"`     // Статус лечения раненых
}
