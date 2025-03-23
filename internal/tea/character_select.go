package teaapp

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type CharacterSelectModel struct {
	app       *App
	cursor    int
	fromMenu  tea.Model
	charCount int
}

func NewCharacterSelectModel(app *App, fromMenu tea.Model) tea.Model {
	return &CharacterSelectModel{
		app:       app,
		fromMenu:  fromMenu,
		cursor:    0,
		charCount: len(app.AllCharacters()),
	}
}

func (m *CharacterSelectModel) Init() tea.Cmd {
	return nil
}

func (m *CharacterSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m.fromMenu, nil
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < m.charCount-1 {
				m.cursor++
			}
		case "enter":
			return NewUsecaseListModelWithChar(m.app, m.cursor, m), nil
		}
	}
	return m, nil
}

func (m *CharacterSelectModel) View() string {
	s := "Select Character:\n\n"
	chars := m.app.AllCharacters()

	for i, char := range chars {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s += fmt.Sprintf(" %s %d) %s (Power: %d, VIP: %d)\n",
			cursor, i+1, char.Nickname, char.Power, char.VIPLevel)
	}

	s += "\n↑ ↓ to move • Enter to select • q to back"
	return s
}
