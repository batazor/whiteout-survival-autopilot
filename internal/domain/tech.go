package domain

type Tech struct {
	State TechState `yaml:"state"` // Состояние технологий.
}

type TechState struct {
	IsAvailable bool   `yaml:"is_available"` // Доступность технологий.
	TextStatus  string `yaml:"TextStatus"`   // Текстовый статус.
}
