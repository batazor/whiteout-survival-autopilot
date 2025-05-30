package vision

import (
	"strings"

	"github.com/agnivade/levenshtein"
)

func FuzzySubstringMatch(ocrText, target string, maxDistance int) bool {
	text := strings.ToLower(ocrText)
	target = strings.ToLower(target)
	tLen := len(target)

	// Допускаем окна длиной target-1, target, target+1
	for l := tLen - 1; l <= tLen+1 && l <= len(text); l++ {
		if l <= 0 {
			continue
		}

		for i := 0; i <= len(text)-l; i++ {
			window := text[i : i+l]
			if levenshtein.ComputeDistance(window, target) <= maxDistance {
				return true
			}
		}
	}

	return false
}
