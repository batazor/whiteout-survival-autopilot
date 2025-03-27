package teaapp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// runUsecase executes a selected usecase for a specific character
func (a *App) runUsecaseByName(usecaseName string, charIndex int) error {
	chars := a.AllCharacters()
	if charIndex < 0 || charIndex >= len(chars) {
		return fmt.Errorf("character index out of range")
	}
	char := chars[charIndex]

	usecases, err := a.loader.LoadAll(a.ctx)
	if err != nil {
		panic(fmt.Sprintf("❌ failed to load usecases: %v", err))
	}

	var usecase *domain.UseCase
	for _, uc := range usecases {
		if uc.Name == usecaseName {
			usecase = uc
			break
		}
	}
	if usecase == nil {
		return fmt.Errorf("usecase '%s' not found", usecaseName)
	}

	current := a.gameFSM.Current()
	if current != usecase.Node {
		a.logger.Info("FSM force transition to starting node",
			slog.String("from", current),
			slog.String("to", usecase.Node),
		)

		if path := a.gameFSM.FindPath(current, usecase.Node); len(path) > 1 {
			fmt.Print("🧭 FSM Auto-Path: ")
			for i, p := range path {
				if i > 0 {
					fmt.Print(" → ")
				}
				fmt.Printf("%s", p)
			}
			fmt.Println()
		}

		a.gameFSM.ForceTo(usecase.Node)
	}

	fmt.Printf("🎬 Running usecase: %s for character %s (ID: %d)\n", usecase.Name, char.Nickname, char.ID)

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelUsecase = cancel
	defer func() { a.cancelUsecase = nil }()

	a.executor.ExecuteUseCase(ctx, usecase, a.state)

	return nil
}
