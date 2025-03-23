package teaapp

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var menuChoices = []string{
	"Start Bot",
	"View State",
	"Quit",
}

type MenuModel struct {
	app       *App
	cursor    int
	quitting  bool
	outputLog string
}

func NewMenuModel(app *App) MenuModel {
	return MenuModel{
		app:    app,
		cursor: 0,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(menuChoices)-1 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0: // Start Bot: character -> usecase
				return NewCharacterSelectModel(m.app), nil
			case 1: // View state
				m.outputLog = fmt.Sprintf("Accounts: %d", len(m.app.state.Accounts))
			case 2: // Quit
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	if m.quitting {
		return "Bye 👋\n"
	}

	s := "🎮 Whiteout Survival Autopilot\n\n"
	s += "Choose an option:\n\n"

	for i, choice := range menuChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n" + m.outputLog
	s += "\n↑ ↓ to move • Enter to select • q to quit"
	return s
}
