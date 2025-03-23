package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type UseCaseLoader interface {
	LoadAll(ctx context.Context) ([]*domain.UseCase, error)
}

func NewUseCaseLoader(dir string) UseCaseLoader {
	return &useCaseLoader{dir: dir}
}

type useCaseLoader struct {
	dir string
}

func (l *useCaseLoader) LoadAll(ctx context.Context) ([]*domain.UseCase, error) {
	var usecases []*domain.UseCase

	err := filepath.Walk(l.dir, func(path string, info os.FileInfo, err error) error {
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
		uc, err := l.loadOne(path)
		if err != nil {
			log.Printf("Error loading usecase from %s: %v", path, err)
			return nil
		}
		log.Printf("Loaded usecase: %s", uc.Name)
		usecases = append(usecases, uc)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk usecases dir: %w", err)
	}

	return usecases, nil
}

func (l *useCaseLoader) loadOne(configFile string) (*domain.UseCase, error) {
	v := viper.New()
	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var uc domain.UseCase
	if err := v.Unmarshal(&uc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal usecase: %w", err)
	}
	return &uc, nil
}
