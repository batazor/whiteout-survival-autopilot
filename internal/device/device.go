package device

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/analyzer"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
)

type Device struct {
	Name       string
	Profiles   domain.Profiles
	Logger     *slog.Logger
	ADB        adb.DeviceController
	FSM        *fsm.GameFSM
	areaLookup *config.AreaLookup
	Queue      *redis_queue.RedisQueue
	Executor   executor.UseCaseExecutor

	activeProfileIdx int
	activeGamerIdx   int
}

func New(name string, profiles domain.Profiles, log *slog.Logger, areaPath string, rdb *redis.Client) (*Device, error) {
	log.Info("🔧 Инициализация ADB-контроллера")
	controller, err := adb.NewController(log, name)
	if err != nil {
		log.Error("❌ Не удалось создать ADB-контроллер", slog.Any("error", err))
		return nil, err
	}

	areaLookup, err := config.LoadAreaReferences(areaPath)
	if err != nil {
		log.Error("❌ Ошибка загрузки area.json:", "error", err)
		return nil, err
	}

	exec := executor.NewUseCaseExecutor(
		log,
		config.NewTriggerEvaluator(),
		analyzer.NewAnalyzer(areaLookup, log),
		controller,
		areaLookup,
	)

	device := &Device{
		Name:       name,
		Profiles:   profiles,
		Logger:     log,
		ADB:        controller,
		FSM:        fsm.NewGame(log, controller, areaLookup),
		areaLookup: areaLookup,
		Queue:      redis_queue.NewBotQueue(rdb, name),
		Executor:   exec,
	}

	// Load usecases from the directory
	go device.loadUseCases(context.Background(), "./usecases")

	return device, nil
}
