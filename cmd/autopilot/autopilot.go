package main

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/batazor/whiteout-survival-autopilot/internal/bot"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/logger"
	"github.com/batazor/whiteout-survival-autopilot/internal/redis_queue"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ─── Redis ───────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Redis недоступен: %v", err)
	}

	// ─── Логгер ──────────────────────────────────────────────────────────────
	appLogger, err := logger.InitializeLogger("app")
	if err != nil {
		log.Fatalf("❌ Не удалось инициализировать логгер: %v", err)
	}

	// ─── Конфигурация устройств / профилей ───────────────────────────────────
	devicesCfg, err := config.LoadDeviceConfig("./db/devices.yaml", "./db/state.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// ─── Предзагрузка use‑case’ов ────────────────────────────────────────────
	redis_queue.PreloadQueues(ctx, rdb, devicesCfg.AllProfiles(), "./usecases")

	// ─── Запуск устройств и ботов ────────────────────────────────────────────
	var wg sync.WaitGroup

	for _, devCfg := range devicesCfg.Devices {
		wg.Add(1)

		go func(dc domain.Device) { // ← корректный тип
			defer wg.Done()

			devLog := appLogger.With("device", dc.Name)

			dev, err := device.New(dc.Name, dc.Profiles, devLog, "./references/area.json", rdb)
			if err != nil {
				devLog.Error("❌ Ошибка создания устройства", slog.Any("err", err))
				return
			}

			for pIdx, p := range dc.Profiles {
				for gIdx := range p.Gamer {
					select {
					case <-ctx.Done():
						return
					default:
					}

					if err := dev.SwitchTo(ctx, pIdx, gIdx); err != nil {
						devLog.Warn("⚠️ Не удалось переключиться", slog.Any("err", err))
						continue
					}

					g := &p.Gamer[gIdx]
					b := bot.NewBot(dev, g, rdb, devLog.With("gamer", g.Nickname))

					b.Play(ctx) // ← Play ничего не возвращает
					time.Sleep(3 * time.Second)
				}
			}
		}(devCfg)
	}

	wg.Wait()
}
