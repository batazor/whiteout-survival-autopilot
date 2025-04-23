package imagetiles

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	finder "github.com/batazor/whiteout-survival-autopilot/internal/analyzer/findIcon"
	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/events/marvelous_fantasia/domain"
)

// DetectTiles ищет все плитки в зоне "marvelous_fantasia_game_area" с помощью шаблонов.
func DetectTiles(
	screenshotPath string,
	areaLookup *config.AreaLookup,
	logger *slog.Logger,
	iconDir string,
) (domain.Tiles, error) {
	const threshold = 0.98
	var tiles domain.Tiles

	// Регулярка: "marvelous_fantasia_icon_beer_level_2_down_a1b2c.png"
	iconRe := regexp.MustCompile(`^marvelous_fantasia_icon_([a-z]+)_level_(\d+)_`)

	// Путь до иконок
	files, err := os.ReadDir(iconDir)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения директории иконок: %w", err)
	}

	// Игровая зона
	gameRegion, ok := areaLookup.Get("marvelous_fantasia_game_area")
	if !ok {
		return nil, fmt.Errorf("зона marvelous_fantasia_game_area не найдена")
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := f.Name()
		matches := iconRe.FindStringSubmatch(name)
		if matches == nil {
			continue // не подходит под шаблон
		}

		kind := matches[1]
		level, _ := strconv.Atoi(matches[2])

		iconPath := filepath.Join(iconDir, name)
		boxes, err := finder.FindIconsInRegion(screenshotPath, iconPath, gameRegion.Zone, threshold, logger)
		if err != nil {
			logger.Error("ошибка поиска иконки", slog.String("icon", name), slog.Any("err", err))
			continue
		}

		for _, bbox := range boxes {
			tiles = append(tiles, domain.Tile{
				Kind:  kind,
				Level: level,
				BBox:  bbox,
			})
		}
	}

	return tiles, nil
}
