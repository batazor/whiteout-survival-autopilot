package gift

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func runWOSRewardsLoop(cfg Config) {
	tick := time.NewTicker(cfg.PollEvery)
	defer tick.Stop()

	for {
		if err := pollWOSRewards(cfg); err != nil {
			cfg.Logger.Error("wosrewards poll error", "err", err)
		}
		<-tick.C
	}
}

func pollWOSRewards(cfg Config) error {
	known := loadGiftCodes(cfg.CodesYAML)
	knownMap := make(map[string]bool)
	for _, c := range known.Codes {
		knownMap[strings.ToUpper(c.Name)] = true
	}

	found := []string{}
	c := colly.NewCollector()
	c.OnHTML("h5.font-bold", func(e *colly.HTMLElement) {
		code := strings.TrimSpace(e.Text)
		if code != "" && !knownMap[strings.ToUpper(code)] {
			found = append(found, code)
		}
	})
	err := c.Visit("https://www.wosrewards.com/")
	if err != nil {
		return fmt.Errorf("colly visit failed: %w", err)
	}

	if len(found) == 0 {
		cfg.Logger.Debug("no new codes from wosrewards")
		return nil
	}

	for _, code := range found {
		cfg.Logger.Info("new wosrewards code", "code", code)
		if err := addOrUpdate(cfg.CodesYAML, code, time.Time{}); err != nil {
			cfg.Logger.Error("add code error", "err", err)
			continue
		}
	}
	go runRedeemer(cfg)
	return nil
}

func loadGiftCodes(path string) domain.GiftCodes {
	var gc domain.GiftCodes
	if b, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(b, &gc)
	}
	return gc
}
