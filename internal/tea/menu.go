package teaapp

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	bubblezone "github.com/lrstanley/bubblezone"
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
	zones     *bubblezone.Manager
	help      help.Model
}

func NewMenuModel(app *App) MenuModel {
	return MenuModel{
		app:    app,
		cursor: 0,
		zones:  bubblezone.New(),
		help: func() help.Model {
			h := help.New()
			h.Styles = helpStyle
			return h
		}(),
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
			return m.activateSelected()
		}
	case tea.MouseMsg:
		for i := range menuChoices {
			if m.zones.Get(fmt.Sprintf("menu-%d", i)).InBounds(msg) {
				m.cursor = i
				if msg.Type == tea.MouseLeft {
					return m.activateSelected()
				}
			}
		}
	}

	return m, nil
}

func (m *MenuModel) activateSelected() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // Start Bot: character -> usecase
		return NewUsecaseListModelWithChar(m.app, m.app.CurrentCharacterIndex, m), nil
	case 1: // View state
		m.outputLog = fmt.Sprintf("Accounts: %d", len(m.app.state.Accounts))
	case 2: // Quit
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m MenuModel) View() string {
	if m.quitting {
		return "Bye 👋\n"
	}

	s := "🎮 Whiteout Survival Autopilot\n\n"

	// Show a connected device
	deviceID := m.app.controller.GetActiveDevice()
	if deviceID != "" {
		s += fmt.Sprintf("📱 Connected device: %s\n", deviceID)
	} else {
		s += "⚠️ No connected device\n"
	}

	// Show selected character info
	allChars := m.app.AllCharacters()
	if idx := m.app.CurrentCharacterIndex; idx >= 0 && idx < len(allChars) {
		char := allChars[idx]
		s += fmt.Sprintf("🧍 Character: %s | Power: %d | VIP: %d | Furnace: %d\n",
			char.Nickname, char.Power, char.Vip_Level, char.Buildings.Furnace.Level)
	}

	s += "\nChoose an option:\n\n"

	for i, choice := range menuChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		item := fmt.Sprintf("%s %s", cursor, choice)
		s += m.zones.Mark(fmt.Sprintf("menu-%d", i), item) + "\n"
	}

	s += "\n" + m.outputLog
	s += "\n\n" + m.help.View(keys)
	return m.zones.Scan(s)
}
