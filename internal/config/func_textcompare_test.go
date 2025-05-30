package config

import "testing"

func TestFuzzySubstringMatch(t *testing.T) {
	cases := []struct {
		name             string
		haystack, needle string
		want             bool
	}{
		{"substring", "Completed J", "completed", true},
		{"edit-distance 1 (delete)", "completd", "completed", true},
		{"edit-distance 1 (substitute)", "Complet3d", "completed", true},
		{"mismatch", "not ready", "completed", false},
	}

	for _, tc := range cases {
		tc := tc // pin range variable
		t.Run(tc.name, func(t *testing.T) {
			if got := FuzzySubstringMatch(tc.haystack, tc.needle, 1); got != tc.want {
				t.Errorf("%q vs %q: got %v, want %v", tc.haystack, tc.needle, got, tc.want)
			}
		})
	}
}
