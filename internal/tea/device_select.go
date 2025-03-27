package teaapp

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	bubblezone "github.com/lrstanley/bubblezone"
)

type DeviceSelectModel struct {
	app     *App
	devices []string
	cursor  int
	zones   *bubblezone.Manager
	help    help.Model
}

func NewDeviceSelectModel(app *App, devices []string) tea.Model {
	return &DeviceSelectModel{
		app:     app,
		devices: devices,
		cursor:  0,
		zones:   bubblezone.New(),
		help: func() help.Model {
			h := help.New()
			h.Styles = helpStyle
			return h
		}(),
	}
}

func (m *DeviceSelectModel) Init() tea.Cmd {
	return nil
}

func (m *DeviceSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.devices)-1 {
				m.cursor++
			}
		case "enter":
			m.app.controller.SetActiveDevice(m.devices[m.cursor])
			return NewMenuModel(m.app), nil
		}

	case tea.MouseMsg:
		for i := range m.devices {
			if m.zones.Get(fmt.Sprintf("device-%d", i)).InBounds(msg) {
				m.cursor = i
				if msg.Type == tea.MouseLeft {
					m.app.controller.SetActiveDevice(m.devices[m.cursor])
					return NewMenuModel(m.app), nil
				}
			}
		}
	}
	return m, nil
}

func (m *DeviceSelectModel) View() string {
	s := "📱 Select ADB Device:\n\n"

	for i, device := range m.devices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		item := fmt.Sprintf("%s %s", cursor, device)
		s += m.zones.Mark(fmt.Sprintf("device-%d", i), item) + "\n"
	}

	s += "\n\n" + m.help.View(keys)
	return m.zones.Scan(s)
}
