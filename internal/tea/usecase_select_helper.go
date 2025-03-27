package teaapp

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
)

type usecaseKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Quit   key.Binding
	Select key.Binding // s → popup select
}

func (k usecaseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit}
}

func (k usecaseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Left, k.Right},
		{k.Select},
		{k.Enter, k.Quit},
	}
}

var usecaseKeys = usecaseKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "prev tab"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "next tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "run usecase"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "back"),
	),
	Select: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "select state"),
	),
}

var triggerStatusDescriptions = map[string]string{
	"✅":  "trigger passed",
	"❌":  "trigger not met",
	"⚠️": "trigger eval error",
}

func triggerStatusLegend() string {
	out := "Legend: "
	for _, icon := range []string{"✅", "❌", "⚠️"} {
		out += fmt.Sprintf("%s = %s   ", icon, triggerStatusDescriptions[icon])
	}
	return out
}
