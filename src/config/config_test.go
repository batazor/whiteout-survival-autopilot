package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/batazor/whiteout-survival-autopilot/src/config"
)

func TestLoadAllUseCases(t *testing.T) {
	// Путь к каталогу с YAML-сценариями.
	// Предполагается, что каталог "usecases" находится в корне модуля.
	usecasesDir := "../../usecases"

	// Получаем список файлов в каталоге.
	entries, err := os.ReadDir(usecasesDir)
	if err != nil {
		t.Fatalf("Failed to read usecases directory: %v", err)
	}

	// Перебираем все файлы.
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(usecasesDir, entry.Name())

		t.Run(entry.Name(), func(t *testing.T) {
			uc, err := config.LoadUseCase(context.Background(), filePath)
			if err != nil {
				t.Errorf("Error loading use case %s: %v", entry.Name(), err)
			} else {
				t.Logf("Successfully loaded use case: %s, Node: %s, FinalNode: %s", uc.Name, uc.Node, uc.FinalNode)
			}
		})
	}
}
