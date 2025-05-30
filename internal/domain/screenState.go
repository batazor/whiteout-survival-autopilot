package domain

import (
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
)

type ScreenState struct {
	IsMainMenu bool   `yaml:"isMainMenu"` // Флаг, указывающий, есть ли события в главном меню.
	IsWelcome  bool   `yaml:"isWelcome"`  // Флаг, указывающий, есть ли события приглашения новых выживших.
	IsMainCity string `yaml:"isMainCity"` // Флаг, указывающий, какой экран активен - город или карта мира.

	CurrentState string `yaml:"currentState"` // Заголовок экрана.
	TitleFact    string `yaml:"titleFact"`    // Заголовок экрана, полученный из анализа скриншота.
}

// Reset сбрасывает состояние экрана.
func (s *ScreenState) Reset() {
	s.IsMainMenu = false
	s.IsWelcome = false
	s.IsMainCity = ""
	s.CurrentState = state.StateMainCity
	s.TitleFact = ""
}
