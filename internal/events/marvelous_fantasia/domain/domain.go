package domain

import (
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// Tile описывает одну плитку на экране Marvelous Fantasia
type Tile struct {
	Kind  string      // тип плитки, например "beer", "star" и т.п.
	Level int         // уровень плитки (1, 2, …)
	BBox  domain.BBox // регион плитки (в процентах)
}

// Tiles представляет набор игровых плиток.
type Tiles []Tile

// BBoxes возвращает массив BBox из всех плиток.
func (ts Tiles) BBoxes() []domain.BBox {
	out := make([]domain.BBox, len(ts))
	for i, tile := range ts {
		out[i] = tile.BBox
	}
	return out
}
