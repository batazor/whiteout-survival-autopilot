package device

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain/state"
	"github.com/batazor/whiteout-survival-autopilot/internal/ocrclient"
)

func (d *Device) DetectedGamer(ctx context.Context) (int, int, error) {
	d.Logger.Info("🚀 Определение текущего игрока")

	// 0. Переходим на экран профиля
	d.FSM.ForceTo(state.StateChiefProfile, nil)

	defer func() {
		// 4. Возвращаемся на главный экран
		d.FSM.ForceTo(state.StateMainCity, nil)
	}()

	zone, ok := d.AreaLookup.Get("chief_profile_nickname")
	if !ok {
		d.Logger.Error("GetRegionByName failed",
			slog.String("region", "chief_profile_nickname"),
		)
		return -1, -1, errors.New("не найдено совпадений с никнеймом")
	}

	region := ocrclient.Region{
		X0: zone.Zone.Min.X,
		Y0: zone.Zone.Min.Y,
		X1: zone.Zone.Max.X,
		Y1: zone.Zone.Max.Y,
	}

	// 3. Распознаём никнейм игрока
	fullOCR, fullErr := d.OCRClient.FetchOCR("", []ocrclient.Region{region}) // debugName можно опустить
	if fullErr != nil {
		d.Logger.Error("Full OCR failed", slog.Any("error", fullErr))
		return -1, -1, fmt.Errorf("full OCR failed: %w", fullErr)
	}

	if len(fullOCR) == 0 {
		d.Logger.Warn("⚠️ Не удалось распознать никнейм игрока", slog.String("region", "chief_profile_nickname"))
		return -1, -1, errors.New("не найдено совпадений с никнеймом")
	}

	nicknameParsed := fullOCR[0].Text

	// drop aliance [RLX]batazor -> batazor
	if strings.Contains(nicknameParsed, "]") {
		nicknameParsed = strings.Split(nicknameParsed, "]")[1]
	}

	d.Logger.Info("🟢 Распознан никнейм", slog.String("parsed", nicknameParsed))

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
		return -1, -1, errors.New("не найдено совпадений с никнеймом")
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

func (d *Device) DetectAndSetCurrentGamer(ctx context.Context) (*domain.Gamer, int, int, error) {
	pIdx, gIdx, err := d.DetectedGamer(ctx)
	if err != nil || pIdx < 0 || gIdx < 0 {
		d.Logger.Warn("⚠️ Не удалось определить активного игрока", slog.Any("err", err))
		return nil, -1, -1, err
	}

	// 💾 Сохраняем как текущего
	d.activeProfileIdx = pIdx
	d.activeGamerIdx = gIdx

	active := &d.Profiles[pIdx].Gamer[gIdx]
	d.Logger.Info("🔎 Активный игрок определён", slog.String("nickname", active.Nickname))

	d.FSM.SetCallback(active)

	// Сбрасываем старый стейт
	active.ScreenState.Reset()

	return active, pIdx, gIdx, nil
}
