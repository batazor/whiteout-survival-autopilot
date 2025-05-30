package analyzer

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseTimeDuration распознаёт строки вроде
//   - 42d171612   → 42 дн 17 ч 16 м 12 с
//   - 3d4h30m     → 3 дн 4 ч 30 м
//   - 90m10s      → 1 ч 30 м 10 с
//
// Возвращает time.Duration (0 при ошибке).
func parseTimeDuration(s string) time.Duration {
	clean := strings.ToLower(strings.ReplaceAll(s, " ", ""))

	// 1) Компактный формат 42dHHMMSS
	if m := regexp.MustCompile(`^(\d+)d(\d{2})(\d{2})(\d{2})$`).FindStringSubmatch(clean); m != nil {
		days, _ := strconv.Atoi(m[1])
		hours, _ := strconv.Atoi(m[2])
		mins, _ := strconv.Atoi(m[3])
		secs, _ := strconv.Atoi(m[4])
		return (time.Duration(days)*24*time.Hour +
			time.Duration(hours)*time.Hour +
			time.Duration(mins)*time.Minute +
			time.Duration(secs)*time.Second)
	}

	// 2) Свободный порядок с суффиксами d/h/m/s
	re := regexp.MustCompile(`(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)
	if m := re.FindStringSubmatch(clean); m != nil {
		var days, hours, mins, secs int
		if m[1] != "" {
			days, _ = strconv.Atoi(m[1])
		}
		if m[2] != "" {
			hours, _ = strconv.Atoi(m[2])
		}
		if m[3] != "" {
			mins, _ = strconv.Atoi(m[3])
		}
		if m[4] != "" {
			secs, _ = strconv.Atoi(m[4])
		}
		return (time.Duration(days)*24*time.Hour +
			time.Duration(hours)*time.Hour +
			time.Duration(mins)*time.Minute +
			time.Duration(secs)*time.Second)
	}
	return 0
}
