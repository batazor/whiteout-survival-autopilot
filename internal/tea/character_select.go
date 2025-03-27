package teaapp

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	bubblezone "github.com/lrstanley/bubblezone"

	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

// CharacterSelectModel представляет модель выбора персонажа.
type CharacterSelectModel struct {
	app       *App
	cursor    int
	charCount int
	zones     *bubblezone.Manager
	help      help.Model
}

// NewCharacterSelectModel создает новую модель выбора персонажа.
func NewCharacterSelectModel(app *App) tea.Model {
	// We begin in the main city
	app.gameFSM.ForceTo(fsm.StateMainCity)

	return &CharacterSelectModel{
		app:       app,
		cursor:    0,
		charCount: len(app.AllCharacters()),
		zones:     bubblezone.New(),
		help: func() help.Model {
			h := help.New()
			h.Styles = helpStyle
			return h
		}(),
	}
}

// Init инициализирует модель.
func (m *CharacterSelectModel) Init() tea.Cmd {
	return nil
}

// Update обрабатывает входящие сообщения и обновляет состояние модели.
func (m *CharacterSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return nil, tea.Quit
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
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			for i := 0; i < m.charCount; i++ {
				zoneID := fmt.Sprintf("char-%d", i)
				if m.zones.Get(zoneID).InBounds(msg) {
					m.cursor = i
					return NewUsecaseListModelWithChar(m.app, i, m), nil
				}
			}
		}
	}

	return m, cmd
}

// View возвращает строковое представление текущего состояния модели.
func (m *CharacterSelectModel) View() string {
	s := "Выберите персонажа:\n\n"
	chars := m.app.AllCharacters()

	for i, char := range chars {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		zoneID := fmt.Sprintf("char-%d", i)
		line := fmt.Sprintf(
			"%s %d) %s (Сила: %d, VIP: %d, Печь: %d)",
			cursor, i+1, char.Nickname, char.Power, char.Vip_Level, char.Buildings.Furnace.Level,
		)
		s += m.zones.Mark(zoneID, line) + "\n"
	}

	s += "\n\n" + m.help.View(keys)
	return m.zones.Scan(s)
}
