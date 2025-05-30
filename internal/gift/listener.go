package gift

import (
	"log/slog"
	"os"
	"time"
)

/*────────── Config ──────────*/
type Config struct {
	UserID       string // ← 1634091876319117312
	DevicesYAML  string
	CodesYAML    string
	PythonDir    string
	PollEvery    time.Duration
	HistoryDepth int
	Logger       *slog.Logger // must be non-nil
}

/*───────── AutoStart ───────*/
func AutoStart(cfg Config) {
	if cfg.Logger == nil {
		panic("gift.AutoStart: Logger is nil")
	}
	if cfg.UserID == "" {
		panic("gift.AutoStart: UserID must be set")
	}
	if cfg.PollEvery == 0 {
		cfg.PollEvery = time.Hour * 24
	}
	if cfg.HistoryDepth <= 0 {
		cfg.HistoryDepth = 10
	}
	bearer := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearer == "" {
		cfg.Logger.Info("twitter listener disabled (TWITTER_BEARER_TOKEN not set)")
		return
	}

	go runTwitterLoop(cfg)
	cfg.Logger.Info("twitter listener started")

	go runWOSRewardsLoop(cfg)
	cfg.Logger.Info("wosrewards.com listener started")
}
