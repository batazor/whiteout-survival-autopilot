package domain

// Buildings представляет собой набор зданий персонажа.
type Buildings struct {
	Queue1 string `yaml:"queue1"`
	Queue2 string `yaml:"queue2"`

	State BuildingState `yaml:"state"`

	Furnace Building `yaml:"furnace"` // Печь.
}

type BuildingState struct {
	Text string `yaml:"text"`
}
