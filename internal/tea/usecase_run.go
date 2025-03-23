package teaapp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
)

// runUsecase executes a selected usecase for a specific character
func (a *App) runUsecase(ucIndex, charIndex int) error {
	chars := a.AllCharacters()
	if charIndex < 0 || charIndex >= len(chars) {
		return fmt.Errorf("character index out of range")
	}
	char := chars[charIndex]

	usecases, err := a.loader.LoadAll(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load usecases: %w", err)
	}
	if ucIndex < 0 || ucIndex >= len(usecases) {
		return fmt.Errorf("usecase index out of range")
	}
	usecase := usecases[ucIndex]

	// Step 1: Print current FSM state
	current := a.gameFSM.Current()
	fmt.Printf("\U0001F4CD Current Screen: %s\n", current)

	// Step 2: Init usecase logger
	ucLogger, err := logger.InitializeLogger(usecase.Name)
	if err != nil {
		a.logger.Error("failed to initialize usecase logger", slog.String("usecase", usecase.Name), slog.Any("error", err))
	} else {
		ucLogger.Info("Usecase Start", slog.String("from", current), slog.String("to", usecase.Node))
	}

	// Step 3: Transition FSM to usecase.Node
	if current != usecase.Node {
		a.logger.Info("FSM transition", slog.String("from", current), slog.String("to", usecase.Node))
		a.gameFSM.ForceTo(usecase.Node)
	}

	// Step 4: Run usecase
	fmt.Printf("\U0001F3AC Running usecase: %s for character %s (ID: %d)\n",
		usecase.Name, char.Nickname, char.ID)

	for i, step := range usecase.Steps {
		fmt.Printf("Step %d/%d: Action: %+v\n", i+1, len(usecase.Steps), step)
		time.Sleep(300 * time.Millisecond)
	}

	// Step 5: Transition to FinalNode
	if usecase.FinalNode != "" && a.gameFSM.Current() != usecase.FinalNode {
		a.gameFSM.ForceTo(usecase.FinalNode)
		a.logger.Info("FSM transition to final node",
			slog.String("from", current),
			slog.String("to", usecase.FinalNode))
	}

	// Step 6: Save updated state
	if err := a.repo.SaveState(a.ctx, a.state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if ucLogger != nil {
		ucLogger.Info("Usecase Finished", slog.String("final_node", usecase.FinalNode))
	}
	return nil
}
