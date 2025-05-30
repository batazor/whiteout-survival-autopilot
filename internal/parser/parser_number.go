package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseNumber converts strings like "6)", "1 234 567", "13,350,651", "4.3M", "4,3M", "2.1K", "900K", "1.0m", "12k", "V VIP 4" to their integer values.
func ParseNumber(s string) int {
	// Remove spaces and trailing punctuations like ')', '.', ',', ':', ';'
	clean := strings.ReplaceAll(s, " ", "")
	clean = strings.TrimRight(clean, ").,:;")
	clean = strings.ToUpper(clean)

	// Regex: integer part (with optional grouping commas) + optional fractional + optional K/M
	re := regexp.MustCompile(`^(\d{1,3}(?:,\d{3})*|\d+)([.,]\d+)?([KM])?$`)
	m := re.FindStringSubmatch(clean)
	if len(m) == 0 {
		// Fallback: extract the first sequence of digits anywhere in the string
		digitRe := regexp.MustCompile(`\d+`)
		ds := digitRe.FindString(clean)
		if ds == "" {
			return 0
		}
		val, _ := strconv.Atoi(ds)
		return val
	}

	// Remove grouping commas from integer part
	intPart := strings.ReplaceAll(m[1], ",", "")
	fracPart := ""
	if m[2] != "" {
		// unify decimal separator
		fracPart = strings.ReplaceAll(m[2], ",", ".")
	}
	numberStr := intPart + fracPart

	// Parse the numeric value
	num, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0
	}

	// Apply multiplier for K/M suffix
	switch m[3] {
	case "M":
		num *= 1_000_000
	case "K":
		num *= 1_000
	}

	return int(num)
}
