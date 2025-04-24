package domain

type Troops struct {
	Infantry Infantry `yaml:"infantry"`
	Lancer   Lancer   `yaml:"lancer"`
	Marksman Marksman `yaml:"marksman"`
}

// Infantry represents the infantry troop type ------------
type Infantry struct {
	State InfantryState `yaml:"state"`
}

type InfantryState struct {
	IsAvailable  bool   `yaml:"isAvailable"`  // Признак доступности пехоты.
	IsTextStatus string `yaml:"isTextStatus"` // Текстовый статус.
}

// Lancer represents the lancer troop type ------------
type Lancer struct {
	State LancerState `yaml:"state"`
}

type LancerState struct {
	IsAvailable  bool   `yaml:"isAvailable"`  // Признак доступности копейщика.
	IsTextStatus string `yaml:"isTextStatus"` // Текстовый статус.
}

// Marksman represents the marksman troop type ------------
type Marksman struct {
	State MarksmanState `yaml:"state"`
}

type MarksmanState struct {
	IsAvailable  bool   `yaml:"isAvailable"`  // Признак доступности стрелка.
	IsTextStatus string `yaml:"isTextStatus"` // Текстовый статус.
}
