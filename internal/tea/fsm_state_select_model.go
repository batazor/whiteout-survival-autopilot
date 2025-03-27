package teaapp

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StateSelectModel struct {
	states   []string
	cursor   int
	onSelect func(state string) tea.Cmd
}

func NewStateSelectModel(states []string, onSelect func(state string) tea.Cmd) StateSelectModel {
	return StateSelectModel{
		states:   states,
		cursor:   0,
		onSelect: onSelect,
	}
}

func (m StateSelectModel) Init() tea.Cmd {
	return nil
}

func (m StateSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return nil, nil // закрыть popup
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.states)-1 {
				m.cursor++
			}
		case "enter":
			return nil, m.onSelect(m.states[m.cursor]) // вызов колбэка
		}
	}
	return m, nil
}

func (m StateSelectModel) View() string {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("63")).
		Render(m.renderList())

	return "\n\n" + lipgloss.PlaceHorizontal(60, lipgloss.Center, s)
}

func (m StateSelectModel) renderList() string {
	out := "🎯 Select FSM state:\n\n"
	for i, state := range m.states {
		cursor := "  "
		if i == m.cursor {
			cursor = "👉"
		}
		out += fmt.Sprintf("%s %s\n", cursor, state)
	}
	return out
}
