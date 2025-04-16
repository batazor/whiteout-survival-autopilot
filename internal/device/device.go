package device

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
)

type Device struct {
	Name       string
	Profiles   domain.Profiles
	Logger     *logger.TracedLogger
	ADB        adb.DeviceController
	FSM        *fsm.GameFSM
	AreaLookup *config.AreaLookup
	rdb        *redis.Client

	activeProfileIdx int
	activeGamerIdx   int
}

func New(name string, profiles domain.Profiles, log *logger.TracedLogger, areaPath string, rdb *redis.Client) (*Device, error) {
	ctx := context.Background()

	log.Info(ctx, "🔧 Инициализация ADB-контроллера")
	controller, err := adb.NewController(log, name)
	if err != nil {
		log.Error(ctx, "❌ Не удалось создать ADB-контроллер", slog.Any("error", err))
		return nil, err
	}

	areaLookup, err := config.LoadAreaReferences(areaPath)
	if err != nil {
		log.Error(ctx, "❌ Ошибка загрузки area.json:", slog.Any("error", err))
		return nil, err
	}

	device := &Device{
		Name:       name,
		Profiles:   profiles,
		Logger:     log,
		ADB:        controller,
		FSM:        fsm.NewGame(ctx, log, controller, areaLookup),
		AreaLookup: areaLookup,
		rdb:        rdb,
	}

	return device, nil
}
