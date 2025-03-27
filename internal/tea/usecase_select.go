package teaapp

import (
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
)

type UsecaseListModel struct {
	app        *App
	cursor     int
	usecases   []*domain.UseCase
	triggerOK  []string // "✅", "❌", or "⚠️"
	selected   *domain.UseCase
	err        error
	quitting   bool
	fromMenu   tea.Model
	tabs       TabModel
	charIndex  int
	lastOutput string
}

func NewUsecaseListModelWithChar(app *App, characterIndex int, from tea.Model) tea.Model {
	model := &UsecaseListModel{
		app:       app,
		cursor:    0,
		fromMenu:  from,
		charIndex: characterIndex,
		tabs:      NewTabs(fsm.AllStates),
	}

	model.reloadUsecases()
	return model
}

func (m *UsecaseListModel) Init() tea.Cmd {
	return nil
}

func (m *UsecaseListModel) reloadUsecases() {
	currentNode := m.tabs.Current()

	all, err := m.app.loader.LoadAll(m.app.ctx)
	filtered := make([]*domain.UseCase, 0, len(all))
	results := make([]string, 0, len(all))

	if err == nil {
		for _, uc := range all {
			if uc.Node != currentNode {
				continue
			}

			triggerStatus := "✅"
			if uc.Trigger != "" {
				ok, err := m.app.evaluator.EvaluateTrigger(uc.Trigger, m.app.state)
				if err != nil {
					m.app.logger.Error("trigger eval error",
						slog.String("usecase", uc.Name),
						slog.String("trigger", uc.Trigger),
						slog.Any("error", err),
					)
					triggerStatus = "⚠️"
				} else if !ok {
					triggerStatus = "❌"
				}
			}

			filtered = append(filtered, uc)
			results = append(results, triggerStatus)
		}
	}

	m.usecases = filtered
	m.triggerOK = results
	m.err = err
}

func (m *UsecaseListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m.fromMenu, nil
		case "left", "right":
			old := m.tabs.Index
			m.tabs, _ = m.tabs.Update(msg)
			if m.tabs.Index != old {
				m.reloadUsecases()
				m.cursor = 0
			}

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
			err := m.app.runUsecase(m.cursor, m.charIndex)
			if err != nil {
				m.app.logger.Error("failed to run usecase",
					slog.String("name", m.selected.Name),
					slog.Any("error", err))
				m.lastOutput = fmt.Sprintf("❌ %s: %v", m.selected.Name, err)
			} else {
				m.lastOutput = fmt.Sprintf("✅ %s executed successfully", m.selected.Name)
			}

			m.reloadUsecases()

			return m, nil
		}
	}
	return m, nil
}

func (m *UsecaseListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Failed to load usecases: %v\nPress q to go back.", m.err)
	}

	s := m.tabs.View() + "\nUsecases:\n"
	for i, uc := range m.usecases {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		triggerStatus := m.triggerOK[i]
		s += fmt.Sprintf(" %s %d) [%s] %s \n", cursor, i+1, triggerStatus, uc.Name)
	}

	// Add a refresh option
	cursor := " "
	if m.cursor == len(m.usecases) {
		cursor = ">"
	}
	s += fmt.Sprintf(" %s %d) 🔄 Refresh (screenshot + re-eval)\n", cursor, len(m.usecases)+1)

	s += fmt.Sprintf("\nTotal: %d usecases\n", len(m.usecases))

	if m.lastOutput != "" {
		var styled string
		if len(m.lastOutput) >= 2 && m.lastOutput[:2] == "✅" {
			styled = outputSuccess.Render(m.lastOutput)
		} else {
			styled = outputError.Render(m.lastOutput)
		}
		s += "\n" + outputBoxStyle.Render(styled)
	}

	s += "\n✅ Passed • ❌ Not Met • ⚠️ Error • ↑ ↓ to move • Enter to select • q to go back"
	return s
}

var outputBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 1).
	MarginTop(1).
	Foreground(lipgloss.Color("15")).
	BorderForeground(lipgloss.Color("241"))

var outputSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
var outputError = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))    // red
