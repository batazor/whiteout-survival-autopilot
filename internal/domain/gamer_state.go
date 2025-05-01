package domain

type ScreenState struct {
	IsMainMenu bool `yaml:"is_main_menu"` // Флаг, указывающий, есть ли события в главном меню.
}
