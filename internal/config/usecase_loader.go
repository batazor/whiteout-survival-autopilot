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
	GetByName(name string) *domain.UseCase
}

type usecaseLoader struct {
	dir     string
	indexed map[string]*domain.UseCase
}

// NewUseCaseLoader returns a loader that reads all .yaml/.yml files under dir.
func NewUseCaseLoader(dir string) UseCaseLoader {
	loader := &usecaseLoader{
		dir:     dir,
		indexed: make(map[string]*domain.UseCase),
	}

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".yaml" && ext != ".yml" {
			return nil
		}

		uc, err := LoadUseCase(context.Background(), path)
		if err != nil {
			log.Printf("error loading usecase %s: %v", path, err)
			return nil
		}

		if filepath.Base(filepath.Dir(path)) == "debug" || uc.TTL == 0 {
			uc.TTL = 1
		}

		loader.indexed[uc.Name] = uc
		return nil
	})

	return loader
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

	// сохранить имя файла в структуре
	uc.SourcePath = configFile

	return &uc, nil
}

func (l *usecaseLoader) LoadAll(ctx context.Context) ([]*domain.UseCase, error) {
	var active []*domain.UseCase

	for _, uc := range l.indexed {
		if filepath.Base(filepath.Dir(uc.SourcePath)) == "debug" || uc.Cron != "" {
			active = append(active, uc)
		}
	}

	return active, nil
}

func (l *usecaseLoader) GetByName(name string) *domain.UseCase {
	if l.indexed == nil {
		return nil
	}
	return l.indexed[name]
}
