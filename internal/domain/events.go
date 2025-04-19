package domain

type Events struct {
	TundraAdventure TundraAdventure `yaml:"tundraAdventure"` // События в тундре
	FrostyFortune   FrostyFortune   `yaml:"frostyFortune"`   // События в ледяной крепости
}

type TundraAdventure struct {
	State TundraAdventureState `yaml:"state"` // Состояние тундры
}

type TundraAdventureState struct {
	// Main City ------------
	IsExist bool `yaml:"isExist"` // Признак существования события

	// Play screen -----------
	Count  int  `yaml:"count"`  // Количество доступных бросков
	IsPlay bool `yaml:"isPlay"` // Признак доступности броска игральной кости

	// Adventurer Drill ------
	IsAdventurerDrillClaimIsExist bool `yaml:"isAdventurerDrillClaimIsExist"` // Признак существования обновления данных
	IsAdventurerDrillClaim        bool `yaml:"isAdventurerDrillClaim"`        // Признак наличия квеста на получение добычи
}

type FrostyFortune struct {
	State FrostyFortuneState `yaml:"state"` // Состояние ледяной крепости
}

type FrostyFortuneState struct {
	IsExist bool `yaml:"isExist"` // Признак существования события
}
