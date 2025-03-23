package app

import (
	"context"
	"fmt"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
	"github.com/batazor/whiteout-survival-autopilot/internal/tui"
)

type App struct {
	ctx       context.Context
	repo      repository.StateRepository
	loader    config.UseCaseLoader
	evaluator config.TriggerEvaluator
	executor  executor.UseCaseExecutor
	gameFSM   *fsm.GameFSM
}

// NewApp constructs our top-level application object.
func NewApp() (*App, error) {
	ctx := context.Background()

	// 1) Load/save state from "db/state.yaml"
	repo := repository.NewFileStateRepository("db/state.yaml")

	// 2) Loads .yaml-based use cases from "usecases"
	loader := config.NewUseCaseLoader("usecases")

	// 3) CEL-based trigger evaluator
	evaluator := config.NewTriggerEvaluator()

	// 4) Executor that will run the scenario steps
	exec := executor.NewUseCaseExecutor()

	// 5) Optionally a game FSM for screen transitions
	gameFSM := fsm.NewGameFSM()

	return &App{
		ctx:       ctx,
		repo:      repo,
		loader:    loader,
		evaluator: evaluator,
		executor:  exec,
		gameFSM:   gameFSM,
	}, nil
}

// Run loads the state, loads usecases, then starts a TUI that includes
// an internal loop for triggers & executing usecases.
func (a *App) Run() error {
	// 1. Load state from repository
	st, err := a.repo.LoadState(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// 2. Load all usecases from "usecases"
	usecases, err := a.loader.LoadAll(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load usecases: %w", err)
	}

	// 3. Hand off control to a Bubble Tea TUI
	//    The TUI can trigger actions, show the user what's happening, etc.
	//    We pass references to everything it needs.
	model := tui.NewModel(st, usecases, a.evaluator, a.executor)
	if err := tui.RunTUI(model); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
