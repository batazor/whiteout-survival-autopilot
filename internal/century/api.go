package century

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	apiURL = "https://wos-giftcode-api.centurygame.com/api/player"
	secret = "tB87#kPtkxqOS2"
)

type PlayerInfo struct {
	Code int `json:"code"`
	Data struct {
		FID          int64  `json:"fid"`
		Nickname     string `json:"nickname"`
		KID          int    `json:"kid"`
		StoveLevel   int    `json:"stove_lv"`
		AvatarImage  string `json:"avatar_image"`
		StoveContent int    `json:"stove_lv_content"` // <-- здесь исправить тип на int
	} `json:"data"`
	Msg     string `json:"msg"`
	ErrCode string `json:"err_code"`
}

func FetchPlayerInfo(fid int) (*PlayerInfo, error) {
	timestamp := time.Now().UnixMilli()
	payload := fmt.Sprintf("fid=%d&time=%d", fid, timestamp)

	// Generate MD5 sign
	hash := md5.Sum([]byte(payload + secret))
	sign := hex.EncodeToString(hash[:])

	// Final POST body
	form := url.Values{}
	form.Set("sign", sign)
	form.Set("fid", strconv.FormatInt(int64(fid), 10))
	form.Set("time", strconv.FormatInt(timestamp, 10))

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var player PlayerInfo
	if err := json.Unmarshal(body, &player); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	if player.Code != 0 {
		return nil, fmt.Errorf("API error: %s", player.Msg)
	}

	return &player, nil
}
