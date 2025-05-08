package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// UseCaseLoader knows how to scan a directory and load all YAML usecases with hot-reload support.
type UseCaseLoader interface {
	LoadAll(ctx context.Context) ([]*domain.UseCase, error)
	GetByName(name string) *domain.UseCase
	Reload(ctx context.Context) error
	Watch(ctx context.Context) error
}

type usecaseLoader struct {
	dir     string
	indexed map[string]*domain.UseCase
	mu      sync.RWMutex
	watcher *fsnotify.Watcher
}

// NewUseCaseLoader returns a loader that reads all .yaml/.yml files under dir.
func NewUseCaseLoader(dir string) UseCaseLoader {
	loader := &usecaseLoader{
		dir:     dir,
		indexed: make(map[string]*domain.UseCase),
	}
	_ = loader.Reload(context.Background())
	return loader
}

func (l *usecaseLoader) reloadIndex(ctx context.Context) error {
	indexed := make(map[string]*domain.UseCase)

	err := filepath.Walk(l.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".yaml" && ext != ".yml" {
			return nil
		}

		uc, err := LoadUseCase(ctx, path)
		if err != nil {
			log.Printf("error loading usecase %s: %v", path, err)
			return nil
		}

		if filepath.Base(filepath.Dir(path)) == "debug" || uc.TTL == 0 {
			uc.TTL = 1
		}

		indexed[uc.Name] = uc
		return nil
	})

	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.indexed = indexed
	return nil
}

// Reload пересканирует директорию и обновляет список usecase-ов.
func (l *usecaseLoader) Reload(ctx context.Context) error {
	return l.reloadIndex(ctx)
}

// Watch включает отслеживание изменений директорий и файлов usecase-ов.
// При изменении .yaml/.yml файлов автоматически обновляет кэш.
func (l *usecaseLoader) Watch(ctx context.Context) error {
	if l.watcher != nil {
		return fmt.Errorf("already watching")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	l.watcher = watcher

	// Рекурсивно следим за всеми поддиректориями
	err = filepath.Walk(l.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Ext(event.Name) == ".yaml" || filepath.Ext(event.Name) == ".yml" {
					log.Printf("[UseCaseLoader] Change detected: %s (%s), reloading...", event.Name, event.Op)
					_ = l.Reload(ctx)
				}
				// Если появилась новая директория — начинаем за ней следить
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						_ = watcher.Add(event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[UseCaseLoader] FSNotify error: %v", err)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
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
	l.mu.RLock()
	defer l.mu.RUnlock()
	var active []*domain.UseCase

	for _, uc := range l.indexed {
		if filepath.Base(filepath.Dir(uc.SourcePath)) == "debug" || uc.Cron != "" {
			active = append(active, uc)
		}
	}

	return active, nil
}

func (l *usecaseLoader) GetByName(name string) *domain.UseCase {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.indexed == nil {
		return nil
	}
	return l.indexed[name]
}
