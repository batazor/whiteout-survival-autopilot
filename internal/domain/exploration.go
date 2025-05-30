package domain

// Exploration описывает уровень исследования мира.
type Exploration struct {
	Level int              `yaml:"level"`
	State ExplorationState `yaml:"state"`

	IsNotify bool `yaml:"isNotify"` // Признак необходимости уведомления о состоянии исследования.
}

// ExplorationState описывает состояние исследования мира.
type ExplorationState struct {
	IsClaimActive bool `yaml:"isClaimActive"` // Признак доступности кнопки "Забрать".

	MyPower      int    `yaml:"myPower"`      // Мощь игрока.
	EnemyPower   int    `yaml:"enemyPower"`   // Мощь противника.
	BattleStatus string `yaml:"battleStatus"` // Состояние боя (например, "victory", "defeat").
}
