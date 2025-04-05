package main

import (
	"context"
	"log"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// TODO: use redis queue
	_ = &redis_queue.RedisQueue{
		Rdb: rdb,
		Key: "bot:usecase_queue",
	}

	appLogger, err := logger.InitializeLogger("app")
	if err != nil {
		log.Fatalf("❌ Не удалось инициализировать логгер: %v", err)
	}

	devicesCfg, err := config.LoadDeviceConfig("./db/devices.yaml", "./db/state.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	var wg sync.WaitGroup

	for _, dev := range devicesCfg.Devices {
		wg.Add(1)

		deviceLogger := appLogger.With("device", dev.Name)

		go func(devName string, profiles domain.Profiles, log *slog.Logger) {
			defer wg.Done()

			d, err := device.New(devName, profiles, log, "./references/area.json")
			if err != nil {
				log.Error("❌ Ошибка создания девайса", "error", err)
				return
			}

			d.Start(ctx)
		}(dev.Name, dev.Profiles, deviceLogger)
	}

	wg.Wait()
}
