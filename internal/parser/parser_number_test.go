package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"123", 123},
		{"1 234", 1234},
		{"1 234 567", 1234567},
		{"4.3M", 4300000},
		{"2.1K", 2100},
		{"900K", 900000},
		{"1.0m", 1000000},
		{"12k", 12000},
		{"invalid", 0},
		{"13,350,651", 13350651},
		{"6)", 6},
		{"V VIP 4", 4},
		{"11,428,826", 11428826},
		{"", 0},
	}

	for _, tt := range tests {
		got := ParseNumber(tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}
