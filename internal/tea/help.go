package teaapp

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = help.Styles{
	ShortKey:  lipgloss.NewStyle().Foreground(lipgloss.Color("241")), // сероватый
	ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color("111")), // голубой
	FullKey:   lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	FullDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color("111")),
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "enter"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "exit"),
	),
}
