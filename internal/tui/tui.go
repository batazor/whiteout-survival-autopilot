package tui

import (
	"fmt"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	state     *domain.State
	usecases  []*domain.UseCase
	evaluator config.TriggerEvaluator
	executor  executor.UseCaseExecutor

	cursor   int // which usecase is selected
	quitting bool
}

func NewModel(
	st *domain.State,
	usecases []*domain.UseCase,
	eval config.TriggerEvaluator,
	exec executor.UseCaseExecutor,
) model {
	return model{
		state:     st,
		usecases:  usecases,
		evaluator: eval,
		executor:  exec,
	}
}

// Init is called when the program starts; return any initial command.
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages (keyboard, etc.)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.usecases)-1 {
				m.cursor++
			}
		case "enter":
			// Run the selected use case
			if len(m.usecases) > 0 {
				uc := m.usecases[m.cursor]
				triggered, err := m.evaluator.EvaluateTrigger(uc.Trigger, m.state.ToMap())
				if err != nil {
					fmt.Println("Trigger Eval Error:", err)
				} else if triggered {
					m.executor.ExecuteUseCase(uc)
				} else {
					fmt.Printf("Trigger not satisfied for usecase: %s\n", uc.Name)
				}
			}
		}
	}
	return m, nil
}

// View draws the UI
func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	s := "Usecases (press up/down to move, enter to run, 'q' to quit):\n\n"

	for i, uc := range m.usecases {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %d. %s\n", cursor, i+1, uc.Name)
	}

	return s + "\n"
}

// RunTUI is a helper to run the Bubble Tea program
func RunTUI(initialModel model) error {
	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}
