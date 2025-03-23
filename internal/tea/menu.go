package teaapp

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var menuChoices = []string{
	"List Usecases",
	"List Characters",
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
			case 0:
				return NewCharacterSelectModel(m.app, m), nil
			case 3:
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
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // selected
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n" + m.outputLog
	s += "\n↑ ↓ to move • Enter to select • q to quit"
	return s
}
