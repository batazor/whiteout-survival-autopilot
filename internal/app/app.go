package app

import (
	"context"
	"fmt"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
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

func NewApp() (*App, error) {
	ctx := context.Background()

	return &App{
		ctx:       ctx,
		repo:      repository.NewFileStateRepository("db/state.yaml"),
		loader:    config.NewUseCaseLoader("usecases"),
		evaluator: config.NewTriggerEvaluator(),
		executor:  executor.NewUseCaseExecutor(),
		gameFSM:   fsm.NewGameFSM(),
	}, nil
}

func (a *App) Run() error {
	// Load current state from disk
	st, err := a.repo.LoadState(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Load all YAML-defined usecases
	usecases, err := a.loader.LoadAll(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to load usecases: %w", err)
	}

	// Launch the Bubble Tea TUI
	model := tui.NewModel(st, usecases, a.evaluator, a.executor)
	if err := tui.RunTUI(model); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
