package teaapp

import tea "github.com/charmbracelet/bubbletea"

type TabModel struct {
	Options []string
	Index   int
}

func NewTabs(options []string) TabModel {
	return TabModel{Options: options}
}

func (t TabModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if t.Index > 0 {
				t.Index--
			}
		case "right":
			if t.Index < len(t.Options)-1 {
				t.Index++
			}
		}
	}
	return t, nil
}

func (t TabModel) Current() string {
	if len(t.Options) == 0 {
		return ""
	}
	return t.Options[t.Index]
}

func (t TabModel) View() string {
	s := ""
	for i, tab := range t.Options {
		if i == t.Index {
			s += "[" + tab + "] "
		} else {
			s += tab + " "
		}
	}
	return s + "\n"
}
