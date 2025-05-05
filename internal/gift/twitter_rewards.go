package gift

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/g8rswimmer/go-twitter/v2"
	"gopkg.in/yaml.v3"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

/*──────── polling loop ─────*/
func runTwitterLoop(cfg Config) {
	bearer := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearer == "" {
		cfg.Logger.Info("twitter listener disabled (TWITTER_BEARER_TOKEN not set)")
		return
	}

	api := newClient(bearer)
	tick := time.NewTicker(cfg.PollEvery)
	defer tick.Stop()

	for {
		if err := pollOnceTwitter(context.Background(), api, cfg); err != nil {
			cfg.Logger.Error("twitter poll error", "err", err)
		}
		<-tick.C
	}
}

func pollOnceTwitter(ctx context.Context, api *client, cfg Config) error {
	tweets, err := api.fetchLastTweets(ctx, cfg.UserID, cfg.HistoryDepth)
	if err != nil {
		return err
	}

	// (?i) — регистронезависимый,  (?: … ) — необязательная группа
	reCode := regexp.MustCompile(`(?i)(?:gift\s*)?code:\s*([A-Za-z0-9]{6,20})`)
	reUntil := regexp.MustCompile(`(?i)valid\s+until:\s*([A-Za-z]+\s+\d{1,2},\s+\d{2}:\d{2})`)

	changed := false
	for i := len(tweets) - 1; i >= 0; i-- {
		txt := tweets[i]

		if m := reCode.FindStringSubmatch(txt); len(m) != 0 {
			code := strings.TrimSpace(m[1])
			exp := parseUntil(reUntil.FindStringSubmatch(txt))

			if needAdd(cfg.CodesYAML, code, exp) {
				cfg.Logger.Info(fmt.Sprintf("new twitter code: %s (expires %v)", code, exp))
				if err := addOrUpdate(cfg.CodesYAML, code, exp); err != nil {
					cfg.Logger.Error("add code error", "err", err)
					continue
				}
				changed = true
			}
		}
	}
	if changed {
		go runRedeemer(cfg)
	}
	return nil
}

func newClient(bearer string) *client {
	return &client{
		api: &twitter.Client{
			Authorizer: bearerAuth{token: bearer},
			Client:     http.DefaultClient,
			Host:       "https://api.twitter.com",
		},
	}
}

func (c *client) fetchLastTweets(ctx context.Context, userID string, limit int) ([]string, error) {
	resp, err := c.api.UserTweetTimeline(ctx, userID, twitter.UserTweetTimelineOpts{
		MaxResults: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(resp.Raw.Tweets))
	for _, tw := range resp.Raw.Tweets {
		out = append(out, tw.Text)
	}
	return out, nil
}

/*──────── twitter client ────*/
type client struct{ api *twitter.Client }

type bearerAuth struct{ token string }

func (b bearerAuth) Add(r *http.Request) { r.Header.Set("Authorization", "Bearer "+b.token) }

/*──────── helpers ──────────*/
func parseUntil(m []string) time.Time {
	if len(m) != 2 {
		return time.Time{}
	}
	t, err := time.ParseInLocation("January 2, 15:04", m[1], time.UTC)
	if err != nil {
		return time.Time{}
	}
	exp := time.Date(time.Now().UTC().Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	if exp.Before(time.Now().UTC()) {
		exp = exp.AddDate(1, 0, 0)
	}
	return exp
}

func needAdd(path, code string, exp time.Time) bool {
	var gc domain.GiftCodes
	if b, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(b, &gc)
	}
	for _, c := range gc.Codes {
		if strings.EqualFold(c.Name, code) {
			old, _ := time.Parse(time.RFC3339, c.Expires)
			return exp.After(old)
		}
	}
	return true
}

func addOrUpdate(path, code string, exp time.Time) error {
	var gc domain.GiftCodes
	if b, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(b, &gc)
	}
	for i := range gc.Codes {
		if strings.EqualFold(gc.Codes[i].Name, code) {
			gc.Codes[i].Expires = exp.UTC().Format(time.RFC3339)
			return saveYAML(path, &gc)
		}
	}
	gc.Codes = append(gc.Codes, domain.GiftCode{
		Name:    code,
		Expires: exp.UTC().Format(time.RFC3339),
		UserFor: map[string]string{},
	})
	return saveYAML(path, &gc)
}

func saveYAML(path string, data any) error {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	b, _ := yaml.Marshal(data)
	return os.WriteFile(path, b, 0o644)
}

/*──────── Redeemer ─────────*/
var mu sync.Mutex

func runRedeemer(cfg Config) {
	mu.Lock()
	defer mu.Unlock()

	cfg.Logger.Info("Redeemer run (twitter)")
	RunRedeemer(RedeemConfig{
		DevicesYAML: cfg.DevicesYAML,
		CodesYAML:   cfg.CodesYAML,
		PythonDir:   cfg.PythonDir,
	})
}
