package teaapp

import (
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type UsecaseListModel struct {
	app       *App
	cursor    int
	usecases  []*domain.UseCase
	triggerOK []string // "✅", "❌", or "⚠️"
	selected  *domain.UseCase
	err       error
	quitting  bool
	fromMenu  tea.Model
	charIndex int
}

func NewUsecaseListModelWithChar(app *App, characterIndex int, from tea.Model) tea.Model {
	model := &UsecaseListModel{
		app:       app,
		cursor:    0,
		fromMenu:  from,
		charIndex: characterIndex,
	}

	model.reloadUsecases()
	return model
}

func (m *UsecaseListModel) Init() tea.Cmd {
	return nil
}

func (m *UsecaseListModel) reloadUsecases() {
	ucs, err := m.app.loader.LoadAll(m.app.ctx)
	results := make([]string, len(ucs))

	if err == nil {
		for i, uc := range ucs {
			if uc.Trigger == "" {
				results[i] = "✅"
				continue
			}
			ok, err := m.app.evaluator.EvaluateTrigger(uc.Trigger, m.app.state)
			if err != nil {
				m.app.logger.Error("trigger eval error",
					slog.String("usecase", uc.Name),
					slog.String("trigger", uc.Trigger),
					slog.Any("error", err),
				)
				results[i] = "⚠️"
			} else if ok {
				results[i] = "✅"
			} else {
				results[i] = "❌"
			}
		}
	}

	m.usecases = ucs
	m.triggerOK = results
	m.err = err
}

func (m *UsecaseListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m.fromMenu, nil

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.usecases) {
				m.cursor++
			}
		case "enter":
			if m.cursor == len(m.usecases) {
				// Refresh option selected
				m.app.UpdateStateFromScreenshot(m.app.gameFSM.Current())
				m.reloadUsecases()
				return m, nil
			}

			m.selected = m.usecases[m.cursor]
			if err := m.app.runUsecase(m.cursor, m.charIndex); err != nil {
				m.app.logger.Error("failed to run usecase",
					slog.String("name", m.selected.Name),
					slog.Any("error", err))
			}
			return m.fromMenu, nil
		}
	}
	return m, nil
}

func (m *UsecaseListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Failed to load usecases: %v\nPress q to go back.", m.err)
	}

	s := "Usecases:\n"
	for i, uc := range m.usecases {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		triggerStatus := m.triggerOK[i]
		s += fmt.Sprintf(" %s %d) [%s] %s (%s → %s)\n",
			cursor, i+1, triggerStatus, uc.Name, uc.Node, uc.FinalNode)
	}

	// Add refresh option
	cursor := " "
	if m.cursor == len(m.usecases) {
		cursor = ">"
	}
	s += fmt.Sprintf(" %s %d) 🔄 Refresh (screenshot + re-eval)\n", cursor, len(m.usecases)+1)

	s += fmt.Sprintf("\nTotal: %d usecases\n", len(m.usecases))
	s += "\n✅ Passed • ❌ Not Met • ⚠️ Error • ↑ ↓ to move • Enter to select • q to go back"
	return s
}
