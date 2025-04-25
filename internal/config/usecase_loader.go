package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// UseCaseLoader knows how to scan a directory and load all YAML usecases.
type UseCaseLoader interface {
	LoadAll(ctx context.Context) ([]*domain.UseCase, error)
}

// NewUseCaseLoader returns a loader that reads all .yaml/.yml files under dir.
func NewUseCaseLoader(dir string) UseCaseLoader {
	return &usecaseLoader{dir: dir}
}

// LoadUseCase reads a single YAML usecase from disk into a domain.UseCase.
func LoadUseCase(ctx context.Context, configFile string) (*domain.UseCase, error) {
	v := viper.New()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read usecase file %s: %w", configFile, err)
	}

	var uc domain.UseCase
	if err := v.Unmarshal(&uc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal usecase %s: %w", configFile, err)
	}

	return &uc, nil
}

type usecaseLoader struct {
	dir string
}

func (l *usecaseLoader) LoadAll(ctx context.Context) ([]*domain.UseCase, error) {
	var usecases []*domain.UseCase

	// Путь до папки debug
	debugPath := filepath.Join(l.dir, "debug")

	hasDebugFiles := false
	_ = filepath.Walk(debugPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			hasDebugFiles = true
			return filepath.SkipDir // прекращаем после первого найденного
		}
		return nil
	})

	// Если есть файлы в debug, грузим только их
	searchDir := l.dir
	if hasDebugFiles {
		searchDir = debugPath
	}

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		uc, err := LoadUseCase(ctx, path)
		if err != nil {
			log.Printf("error loading usecase %s: %v", path, err)
			return nil
		}

		if filepath.Base(filepath.Dir(path)) == "debug" || uc.TTL == 0 {
			uc.TTL = 1 // Устанавливаем TTL в 1, чтобы usecase не попал в очередь
		}

		usecases = append(usecases, uc)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed walking usecases dir: %w", err)
	}

	return usecases, nil
}
