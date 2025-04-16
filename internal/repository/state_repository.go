package repository

import (
	"context"
	"fmt"
	"os"
	"sort"

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
	sort.Sort(s.Gamers)

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

	found := false
	for i, g := range state.Gamers {
		if g.ID == gamer.ID {
			state.Gamers[i] = *gamer
			found = true
			break
		}
	}

	if !found {
		state.Gamers = append(state.Gamers, *gamer)
	}

	return r.SaveState(ctx, state)
}
