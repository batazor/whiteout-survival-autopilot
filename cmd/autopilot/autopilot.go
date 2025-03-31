package main

import (
	"context"
	"log"
	"log/slog"
	"sync"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
)

func main() {
	ctx := context.Background()

	appLogger, err := logger.InitializeLogger("app")
	if err != nil {
		log.Fatalf("❌ Не удалось инициализировать логгер: %v", err)
	}

	devicesCfg, err := config.LoadDeviceConfig("./db/devices.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	areaLookup, err := config.LoadAreaReferences("./references/area.json")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки area.json: %v", err)
	}

	var wg sync.WaitGroup

	for _, dev := range devicesCfg.Devices {
		wg.Add(1)

		deviceLogger := appLogger.With("device", dev.Name)

		go func(devName string, profiles []domain.Profile, log *slog.Logger) {
			defer wg.Done()

			d, err := device.New(devName, profiles, log, areaLookup)
			if err != nil {
				log.Error("❌ Ошибка создания девайса", "error", err)
				return
			}

			d.Start(ctx)
		}(dev.Name, dev.Profiles, deviceLogger)
	}

	wg.Wait()
}
