package device

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/batazor/whiteout-survival-autopilot/internal/fsm"
	"github.com/batazor/whiteout-survival-autopilot/internal/vision"
)

func (d *Device) DetectedGamer(ctx context.Context, imagePath string) (int, int, error) {
	d.Logger.Info("🚀 Определение текущего игрока")

	// 0. Переходим на экран профиля
	d.FSM.ForceTo(fsm.StateChiefProfile)

	// 1. Делаем скриншот экрана профиля
	_, err := d.ADB.Screenshot(imagePath)
	if err != nil {
		d.Logger.Error("❌ Не удалось сделать скриншот для определения игрока", slog.Any("err", err))
		return -1, -1, err
	}

	// 2. Определяем активного игрока через OCR
	zones, ok := d.areaLookup.Get("chief_profile_nickname")
	if !ok {
		return -1, -1, errors.New("❌ зона 'chief_profile_nickname' не найдена в area.json")
	}

	// 3. Распознаём никнейм игрока
	nicknameRaw, err := vision.ExtractTextFromRegion(imagePath, zones.Zone, "gamer_detected")
	if err != nil {
		return -1, -1, err
	}
	nicknameParsed := strings.ToLower(strings.TrimSpace(nicknameRaw))

	// drop aliance [RLX]batazor -> batazor
	if strings.Contains(nicknameParsed, "]") {
		nicknameParsed = strings.Split(nicknameParsed, "]")[1]
	}

	d.Logger.Info("🟢 Распознан никнейм", slog.String("raw", nicknameRaw), slog.String("parsed", nicknameParsed))

	type matchInfo struct {
		profileIdx int
		gamerIdx   int
		score      int
	}

	var matches []matchInfo

	for pIdx, profile := range d.Profiles {
		for gIdx, gamer := range profile.Gamer {
			expected := strings.ToLower(strings.TrimSpace(gamer.Nickname))
			if matched := fuzzy.RankMatch(expected, nicknameParsed); matched != -1 {
				matches = append(matches, matchInfo{pIdx, gIdx, matched})
			}
		}
	}

	if len(matches) == 0 {
		d.Logger.Warn("⚠️ Никнейм не найден по нечёткому совпадению", slog.String("parsed", nicknameParsed))
		return -1, -1, nil
	}

	// Находим наилучшее совпадение (с самым низким score)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score < matches[j].score
	})
	best := matches[0]

	d.Logger.Info("✅ Найден игрок",
		slog.Int("profileIdx", best.profileIdx),
		slog.Int("gamerIdx", best.gamerIdx),
		slog.Int("score", best.score),
	)

	return best.profileIdx, best.gamerIdx, nil
}
