package domain

type ScreenState struct {
	IsMainMenu bool `yaml:"isMainMenu"` // Флаг, указывающий, есть ли события в главном меню.
	IsWelcome  bool `yaml:"isWelcome"`  // Флаг, указывающий, есть ли события приглашения новых выживших.

	CurrentState string `yaml:"currentState"` // Заголовок экрана.
	TitleFact    string `yaml:"titleFact"`    // Заголовок экрана, полученный из анализа скриншота.
}
