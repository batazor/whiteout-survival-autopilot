package device

import (
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/adb"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
)

type Device struct {
	Name             string
	Profiles         domain.Profiles
	Logger           *slog.Logger
	ADB              adb.DeviceController
	FSM              *fsm.GameFSM
	AreaLookup       *config.AreaLookup
	rdb              *redis.Client
	triggerEvaluator config.TriggerEvaluator
	OCRClient        *ocrclient.Client

	activeProfileIdx int
	activeGamerIdx   int
}

func New(deviceId string, profiles domain.Profiles, log *slog.Logger, areaPath string, rdb *redis.Client,
	triggerEvaluator config.TriggerEvaluator) (*Device, error) {

	log.Info("🔧 Инициализация ADB-контроллера")
	controller, err := adb.NewController(log, deviceId)
	if err != nil {
		log.Error("❌ Не удалось создать ADB-контроллер", slog.Any("error", err))
		return nil, err
	}

	areaLookup, err := config.LoadAreaReferences(areaPath)
	if err != nil {
		log.Error("❌ Ошибка загрузки area.json:", "error", err)
		return nil, err
	}

	device := &Device{
		Name:             deviceId,
		Profiles:         profiles,
		Logger:           log,
		ADB:              controller,
		AreaLookup:       areaLookup,
		rdb:              rdb,
		triggerEvaluator: triggerEvaluator,
		OCRClient:        ocrclient.NewClient(deviceId, log),
	}

	// Инициализация FSM
	device.FSM = fsm.NewGame(log, controller, areaLookup, triggerEvaluator, device.ActiveGamer(), device.OCRClient)

	return device, nil
}
