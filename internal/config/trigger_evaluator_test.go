package config

import (
	"testing"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

func TestCompareTextCEL(t *testing.T) {
	eval := NewTriggerEvaluator()
	gamer := &domain.Gamer{} // реальные данные здесь не нужны

	tests := []struct {
		expr string
		want bool
	}{
		{`compareText("Completed", "Completed J")`, true},
		{`compareText("Completed", "completd")`, true},
		{`compareText("Idle", "Idl")`, true},
	}

	for _, tc := range tests {
		got, err := eval.EvaluateTrigger(tc.expr, gamer)
		if err != nil {
			t.Fatalf("expr %q: unexpected error: %v", tc.expr, err)
		}
		if got != tc.want {
			t.Errorf("expr %q: got %v, want %v", tc.expr, got, tc.want)
		}
	}
}
