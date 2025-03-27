package executor_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
	"github.com/batazor/whiteout-survival-autopilot/internal/executor"
)

// mockEvaluator считает вызовы и возвращает true только два раза
type mockEvaluator struct {
	counter int
}

func (m *mockEvaluator) EvaluateTrigger(expr string, state *domain.State) (bool, error) {
	if m.counter < 2 {
		m.counter++
		return true, nil
	}
	return false, nil
}

type noopAnalyzer struct{}

func (a *noopAnalyzer) AnalyzeAndUpdateState(imagePath string, state *domain.State, rules []domain.AnalyzeRule) (*domain.State, error) {
	return state, nil
}

type noopADB struct{}

func (a *noopADB) Screenshot(path string) error {
	return nil
}

func TestLoopExecution(t *testing.T) {
	require := require.New(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	state := &domain.State{}
	evaluator := &mockEvaluator{}
	analyzer := &noopAnalyzer{}
	adb := &noopADB{}

	exec := executor.NewUseCaseExecutor(logger, evaluator, analyzer, adb)

	usecase := &domain.UseCase{
		Name:    "Test Loop",
		Node:    "loop_node",
		Trigger: "true",
		Steps: domain.Steps{
			{
				Action:  "loop",
				Trigger: "always_true",
				Steps: domain.Steps{
					{
						Click: "test_button",
					},
				},
			},
		},
	}

	exec.ExecuteUseCase(context.TODO(), usecase, state)

	require.Equal(2, evaluator.counter, "loop should evaluate trigger exactly 2 times")
}
