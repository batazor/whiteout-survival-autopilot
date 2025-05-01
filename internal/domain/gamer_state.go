package domain

type ScreenState struct {
	IsMainMenu bool `yaml:"isMainMenu"` // Флаг, указывающий, есть ли события в главном меню.
}
