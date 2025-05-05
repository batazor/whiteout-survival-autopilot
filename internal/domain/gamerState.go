package domain

type ScreenState struct {
	IsMainMenu bool `yaml:"isMainMenu"` // Флаг, указывающий, есть ли события в главном меню.
	IsWelcome  bool `yaml:"isWelcome"`  // Флаг, указывающий, есть ли события приглашения новых выживших.
}
