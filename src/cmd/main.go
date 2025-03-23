package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/src/config"
	"github.com/batazor/whiteout-survival-autopilot/src/executor"
)

func main() {
	ctx := context.Background()
	usecasesDir := "usecases"

	var usecases []*config.UseCase

	// Ищем все YAML-файлы в каталоге usecases.
	err := filepath.Walk(usecasesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		uc, err := config.LoadUseCase(ctx, path)
		if err != nil {
			log.Printf("Ошибка загрузки usecase из %s: %v", path, err)
		} else {
			log.Printf("Загружен usecase: %s", uc.Name)
			usecases = append(usecases, uc)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Ошибка обхода каталога usecases: %v", err)
	}

	// Пример состояния; в реальном приложении состояние будет получаться динамически.
	state := map[string]interface{}{
		"isNewMessage":           true,
		"claimButtonIsGreen":     true,
		"isExplorationAvailable": true,
	}

	// Бесконечный цикл выполнения сценариев
	for {
		for _, uc := range usecases {
			triggered, err := config.EvaluateTrigger(uc.Trigger, state)
			if err != nil {
				log.Printf("Ошибка при оценке триггера для %s: %v", uc.Name, err)
				continue
			}
			if triggered {
				log.Printf("Выполняется usecase: %s", uc.Name)
				executor.ExecuteUseCase(uc)
			} else {
				log.Printf("Триггер не выполнен для usecase: %s", uc.Name)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
