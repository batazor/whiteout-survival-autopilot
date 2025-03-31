package teaapp

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
)

type App struct {
	ctx           context.Context
	repo          repository.StateRepository
	loader        config.UseCaseLoader
	evaluator     config.TriggerEvaluator
	gameFSM       *fsm.GameFSM
	state         *domain.State
	areas         *config.AreaLookup
	rules         config.ScreenAnalyzeRules
	analyzer      *analyzer.Analyzer
	controller    adb.DeviceController
	executor      executor.UseCaseExecutor
	cancelUsecase context.CancelFunc
	logger        *slog.Logger

	CurrentCharacterIndex int
}

func NewApp() (*App, error) {
	ctx := context.Background()

	// Global logger
	appLogger, err := logger.InitializeLogger("app")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize app logger: %w", err)
	}

	// Load area references
	areas, err := config.LoadAreaReferences("references/area.json")
	if err != nil {
		appLogger.Error("failed to load area.json", slog.Any("error", err))
		return nil, fmt.Errorf("failed to load area.json: %w", err)
	}

	// Load screen analysis rules
	rules, err := config.LoadAnalyzeRules("references/analyze.yaml")
	if err != nil {
		appLogger.Error("failed to load analyze.yaml", slog.Any("error", err))
		return nil, fmt.Errorf("failed to load analyze.yaml: %w", err)
	}

	// Init ADB
	adbController, err := InitADBController(appLogger)
	if err != nil {
		return nil, err
	}

	// Initialize analyzerConttroller
	analyzerConttroller := analyzer.NewAnalyzer(areas, appLogger)

	evaluator := config.NewTriggerEvaluator()

	app := &App{
		ctx:        ctx,
		repo:       repository.NewFileStateRepository("db/state.yaml"),
		loader:     config.NewUseCaseLoader("usecases"),
		evaluator:  evaluator,
		gameFSM:    fsm.NewGame(appLogger, adbController, areas),
		areas:      areas,
		rules:      rules,
		controller: adbController,
		logger:     appLogger,
		analyzer:   analyzerConttroller,
		executor:   executor.NewUseCaseExecutor(appLogger, evaluator, analyzerConttroller, adbController, areas),
	}

	// Load saved state
	state, err := app.repo.LoadState(ctx)
	if err != nil {
		appLogger.Error("failed to load state.yaml", slog.Any("error", err))
		return nil, fmt.Errorf("failed to load initial state: %w", err)
	}
	app.state = state

	// FSM callbacks
	app.gameFSM.SetCallback(app)
	app.gameFSM.SetStateGetter(func() *domain.State {
		return app.state
	})

	// Fetch additional player data from Century API
	app.UpdateCharacterInfoFromCentury()

	if err := app.repo.SaveState(ctx, app.state); err != nil {
		appLogger.Error("failed to persist state after analysis", slog.Any("error", err))
	}

	appLogger.Info("App initialized", slog.Int("accounts", len(state.Accounts)))

	return app, nil
}

func (a *App) Run() error {
	devices, err := a.controller.ListDevices()
	if err != nil {
		a.logger.Error("failed to list adb devices", slog.Any("error", err))
		return err
	}

	switch len(devices) {
	case 0:
		return fmt.Errorf("no ADB devices connected")

	case 1:
		a.controller.SetActiveDevice(devices[0])
		a.logger.Info("ADB device selected automatically", slog.String("device", devices[0]))
		return tea.NewProgram(NewCharacterSelectModel(a)).Start()
	default:
		// Multiple devices, prompt user to select
		a.logger.Info("multiple ADB devices found", slog.Int("count", len(devices)))
		return tea.NewProgram(NewDeviceSelectModel(a, devices)).Start()
	}
}

// AllCharacters returns all characters across all accounts
func (a *App) AllCharacters() []domain.Gamer {
	var characters []domain.Gamer
	for _, acc := range a.state.Accounts {
		characters = append(characters, acc.Characters...)
	}
	return characters
}

// UpdateStateFromScreenshot captures, analyzes and saves new state
func (a *App) UpdateStateFromScreenshot(screen string) {
	imagePath := "screenshots/current.png"

	// Capture screenshot
	if err := a.controller.Screenshot(imagePath); err != nil {
		a.logger.Error("failed to capture screenshot", slog.Any("error", err))
		return
	}

	rules, ok := a.rules[screen]
	if !ok {
		a.logger.Warn("no analysis rules found for screen", slog.String("screen", screen))
	}

	newState, err := a.analyzer.AnalyzeAndUpdateState(imagePath, a.state, rules)
	if err != nil {
		a.logger.Error("analysis failed", slog.Any("error", err))
		return
	}
	a.state = newState

	if err := a.repo.SaveState(a.ctx, a.state); err != nil {
		a.logger.Error("failed to save state after analysis", slog.Any("error", err))
	}
}
