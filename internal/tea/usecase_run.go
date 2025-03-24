package teaapp

import (
	"fmt"
	"log/slog"
	"path/filepath"
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

	current := a.gameFSM.Current()
	fmt.Printf("📍 Current Screen: %s\n", current)

	// Usecase logger
	ucLogger, err := logger.InitializeLogger(usecase.Name)
	if err != nil {
		a.logger.Error("failed to init usecase logger",
			slog.String("usecase", usecase.Name),
			slog.Any("error", err),
		)
	}
	if ucLogger != nil {
		ucLogger.Info("Usecase Start", slog.String("from", current), slog.String("to", usecase.Node))
	}

	// --- Step 1: Transition to initial screen if needed ---
	if current != usecase.Node {
		a.logger.Info("FSM force transition to starting node",
			slog.String("from", current),
			slog.String("to", usecase.Node),
		)
		a.gameFSM.ForceTo(usecase.Node)
	}

	// --- Step 2: Run the usecase steps ---
	fmt.Printf("🎬 Running usecase: %s for character %s (ID: %d)\n", usecase.Name, char.Nickname, char.ID)

	for i, step := range usecase.Steps {
		stepInfo := fmt.Sprintf("Step %d/%d → Action: %+v", i+1, len(usecase.Steps), step)
		fmt.Println(stepInfo)
		if ucLogger != nil {
			ucLogger.Info("Executing step", slog.Int("step_number", i+1), slog.Any("step", step))
		}

		switch {
		case step.Click != "":
			fmt.Printf("🖱️ Click: %s\n", step.Click)
			if ucLogger != nil {
				ucLogger.Info("Click action", slog.String("target", step.Click))
			}

			if err := a.controller.ClickRegion(step.Click, a.areas); err != nil {
				fmt.Printf("❌ Failed to click: %v\n", err)
				if ucLogger != nil {
					ucLogger.Error("Click failed", slog.String("region", step.Click), slog.Any("error", err))
				}
			}

			time.Sleep(500 * time.Millisecond)

		case step.Wait > 0:
			fmt.Printf("⏳ Wait: %s\n", step.Wait)
			if ucLogger != nil {
				ucLogger.Info("Wait action", slog.String("duration", step.Wait.String()))
			}
			time.Sleep(step.Wait)

		default:
			fmt.Println("⚠️ Unknown step type or empty step")
			if ucLogger != nil {
				ucLogger.Warn("Unknown step type", slog.Any("step", step))
			}
		}
	}

	// --- Step 3: Transition to final node ---
	if usecase.FinalNode != "" && a.gameFSM.Current() != usecase.FinalNode {
		a.logger.Info("FSM transition to final node",
			slog.String("from", usecase.Node),
			slog.String("to", usecase.FinalNode),
		)
		a.gameFSM.ForceTo(usecase.FinalNode)
	}

	// --- Step 4: Analyze screen after usecase (if screenshot already exists) ---
	afterPath := filepath.Join("screenshots", "after_"+usecase.FinalNode+".png")
	newState, err := a.analyzer.AnalyzeAndUpdateState(afterPath, a.state, usecase.FinalNode)
	if err != nil {
		a.logger.Warn("post-usecase state analysis failed", slog.Any("error", err))
	} else {
		a.state = newState
	}

	// --- Step 5: Save updated state ---
	if err := a.repo.SaveState(a.ctx, a.state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if ucLogger != nil {
		ucLogger.Info("Usecase Finished", slog.String("final_node", usecase.FinalNode))
	}

	return nil
}
