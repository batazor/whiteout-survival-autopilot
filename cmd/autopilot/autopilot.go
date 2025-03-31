package main

import (
	"log"
	"sync"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/device"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func main() {
	devicesCfg, err := config.LoadDeviceConfig("./db/devices.yaml")
	if err != nil {
		log.Fatalf("ошибка загрузки конфигурации: %v", err)
	}

	var wg sync.WaitGroup

	for _, dev := range devicesCfg.Devices {
		wg.Add(1)

		go func(devName string, profiles []domain.Profile) {
			defer wg.Done()
			device.Start(devName, profiles)
		}(dev.Name, dev.Profiles)
	}

	wg.Wait()
}
