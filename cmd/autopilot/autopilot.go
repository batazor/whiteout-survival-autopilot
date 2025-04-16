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
	"github.com/batazor/whiteout-survival-autopilot/internal/repository"
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

	// ─── Хранилище состояния ─────────────────────────────────────────────────
	repo := repository.NewFileStateRepository("./db/state.yaml")

	// ─── Конфигурация устройств / профилей ───────────────────────────────────
	devicesCfg, err := config.LoadDeviceConfig("./db/devices.yaml", "./db/state.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// ─── Предзагрузка use‑case’ов ────────────────────────────────────────────
	redis_queue.PreloadQueues(ctx, rdb, devicesCfg.AllProfiles(), "./usecases")

	// ── Запуск глобального рефиллера задач ───────────────────────────────
	go redis_queue.StartGlobalUsecaseRefiller(ctx, devicesCfg, "./usecases", rdb, appLogger, 5*time.Minute)

	// ─── Инициализация правил анализа экрана ───────────────────────────────────────
	rules, err := config.LoadAnalyzeRules("references/analyze.yaml")
	if err != nil {
		appLogger.Error("❌ Ошибка загрузки правил анализа экрана", slog.Any("err", err))
		return
	}

	// ─── Запуск устройств и ботов ────────────────────────────────────────────
	var wg sync.WaitGroup

	for _, devCfg := range devicesCfg.Devices {
		wg.Add(1)

		go func(dc domain.Device) {
			defer wg.Done()

			devLog := appLogger.With("device", dc.Name)

			dev, err := device.New(dc.Name, dc.Profiles, devLog, "./references/area.json", rdb)
			if err != nil {
				devLog.Error("❌ Ошибка создания устройства", slog.Any("err", err))
				return
			}

			activeGamer, pIdx, gIdx, err := dev.DetectAndSetCurrentGamer(ctx)
			if err != nil || activeGamer == nil {
				devLog.Warn("⚠️ Не удалось определить активного игрока", slog.Any("err", err))
				return
			}

			devLog.Info("▶️ Продолжаем с текущего игрока", slog.Int("pIdx", pIdx), slog.Int("gIdx", gIdx), slog.String("nickname", activeGamer.Nickname))

			for {
				select {
				case <-ctx.Done():
					devLog.Info("🛑 Остановка по контексту")
					return
				default:
				}

				if pIdx >= len(dc.Profiles) {
					pIdx = 0
				}
				if gIdx >= len(dc.Profiles[pIdx].Gamer) {
					pIdx++
					gIdx = 0
					continue
				}

				target := &dc.Profiles[pIdx].Gamer[gIdx]
				if dev.ActiveGamer() == nil || dev.ActiveGamer().ID != target.ID {
					if err := dev.SwitchTo(ctx, pIdx, gIdx); err != nil {
						devLog.Warn("⚠️ Не удалось переключиться", slog.Any("err", err))
						gIdx++
						continue
					}
				}

				b := bot.NewBot(dev, target, rdb, rules, devLog.With("gamer", target.Nickname), repo)
				b.Play(ctx)

				gIdx++
			}
		}(devCfg)
	}

	wg.Wait()
}
