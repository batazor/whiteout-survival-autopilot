package domain

type Arena struct {
	Rank    int `yaml:"rank"`    // Ранг игрока на арене
	MyPower int `yaml:"myPower"` // Мощь игрока

	State ArenaState `yaml:"state"` // Состояние арены (например, "open", "closed", "in_battle").
}

type ArenaState struct {
	IsFreeRefresh       bool `yaml:"isFreeRefresh"`       // Признак доступности бесплатного обновления противника.
	IsAvailableFight    bool `yaml:"isAvailableFight"`    // Признак доступности боя.
	CountAvailableFight int  `yaml:"countAvailableFight"` // Количество доступных боев.

	EnemyPower1 int `yaml:"enemyPower1"` // Мощь первого противника.
	EnemyPower2 int `yaml:"enemyPower2"` // Мощь второго противника.
	EnemyPower3 int `yaml:"enemyPower3"` // Мощь третьего противника.
	EnemyPower4 int `yaml:"enemyPower4"` // Мощь четвертого противника.
	EnemyPower5 int `yaml:"enemyPower5"` // Мощь пятого противника.
}
