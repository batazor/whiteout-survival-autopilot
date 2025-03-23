package teaapp

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

type App struct {
	ctx       context.Context
	repo      repository.StateRepository
	loader    config.UseCaseLoader
	evaluator config.TriggerEvaluator
	executor  executor.UseCaseExecutor
	gameFSM   *fsm.GameFSM
	state     *domain.State
	logger    *slog.Logger
}

func NewApp() (*App, error) {
	ctx := context.Background()

	// Initialize app-wide logger
	appLogger, err := logger.InitializeLogger("app")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize app logger: %w", err)
	}

	app := &App{
		ctx:       ctx,
		repo:      repository.NewFileStateRepository("db/state.yaml"),
		loader:    config.NewUseCaseLoader("usecases"),
		evaluator: config.NewTriggerEvaluator(),
		executor:  executor.NewUseCaseExecutor(),
		gameFSM:   fsm.NewGameFSM(appLogger),
		logger:    appLogger,
	}

	state, err := app.repo.LoadState(ctx)
	if err != nil {
		appLogger.Error("failed to load state.yaml", slog.Any("error", err))
		return nil, fmt.Errorf("failed to load initial state: %w", err)
	}
	app.state = state

	appLogger.Info("App initialized", slog.Int("accounts", len(state.Accounts)))
	return app, nil
}

func (a *App) Run() error {
	model := NewMenuModel(a)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}

// AllCharacters returns all characters across all accounts
func (a *App) AllCharacters() []domain.Gamer {
	var characters []domain.Gamer
	for _, acc := range a.state.Accounts {
		characters = append(characters, acc.Characters...)
	}
	return characters
}
