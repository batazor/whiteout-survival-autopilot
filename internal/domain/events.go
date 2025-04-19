package domain

type Events struct {
	TundraAdventure TundraAdventure `yaml:"tundraAdventure"` // События в тундре
	FrostyFortune   FrostyFortune   `yaml:"frostyFortune"`   // События в ледяной крепости
}

type TundraAdventure struct {
	State TundraAdventureState `yaml:"state"` // Состояние тундры
}

type TundraAdventureState struct {
	Count   int  `yaml:"count"`   // Количество доступных бросков
	IsExist bool `yaml:"isExist"` // Признак существования события
	IsPlay  bool `yaml:"isPlay"`  // Признак доступности броска игральной кости
}

type FrostyFortune struct {
	State FrostyFortuneState `yaml:"state"` // Состояние ледяной крепости
}

type FrostyFortuneState struct {
	IsExist bool `yaml:"isExist"` // Признак существования события
}
