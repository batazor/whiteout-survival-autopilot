package repository

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type StateRepository interface {
	LoadState(ctx context.Context) (*domain.State, error)
	SaveState(ctx context.Context, s *domain.State) error
	SaveGamer(ctx context.Context, gamer *domain.Gamer) error
}

func NewFileStateRepository(filename string) StateRepository {
	return &fileRepo{filename: filename}
}

type fileRepo struct {
	filename string
}

func (r *fileRepo) LoadState(ctx context.Context) (*domain.State, error) {
	data, err := os.ReadFile(r.filename)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	var st domain.State
	if err := yaml.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	return &st, nil
}

func (r *fileRepo) SaveState(ctx context.Context, s *domain.State) error {
	bytes, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	return os.WriteFile(r.filename, bytes, 0o644)
}

func (r *fileRepo) SaveGamer(ctx context.Context, gamer *domain.Gamer) error {
	state, err := r.LoadState(ctx)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Ищем нужного игрока по ID и заменяем
	for accIdx := range state.Accounts {
		for charIdx := range state.Accounts[accIdx].Characters {
			if state.Accounts[accIdx].Characters[charIdx].ID == gamer.ID {
				state.Accounts[accIdx].Characters[charIdx] = *gamer
				return r.SaveState(ctx, state)
			}
		}
	}

	return fmt.Errorf("gamer with ID %d not found", gamer.ID)
}
