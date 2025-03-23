package config

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type TriggerEvaluator interface {
	EvaluateTrigger(expr string, st *domain.State) (bool, error)
}

func NewTriggerEvaluator() TriggerEvaluator {
	return &triggerEvaluator{}
}

type triggerEvaluator struct{}

// EvaluateTrigger compiles the CEL expression (expr) and then evaluates it
// against the data we extract from *domain.State.
func (t *triggerEvaluator) EvaluateTrigger(expr string, st *domain.State) (bool, error) {
	// 1. Declare what variables we’ll expose to CEL.
	//    For example, let's define two variables:
	//    - account_count : how many accounts
	//    - has_accounts  : whether there's at least one account
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("account_count", decls.Int),
			decls.NewVar("has_accounts", decls.Bool),
		),
	)
	if err != nil {
		return false, fmt.Errorf("creating CEL env: %w", err)
	}

	// 2. Compile the expression
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile error: %v", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program creation error: %w", err)
	}

	// 3. Build our data map from the *domain.State
	data := map[string]interface{}{
		"account_count": int64(len(st.Accounts)),
		"has_accounts":  len(st.Accounts) > 0,
	}

	// 4. Evaluate the expression
	out, _, err := prg.Eval(data)
	if err != nil {
		return false, fmt.Errorf("eval error: %w", err)
	}

	// 5. Ensure the expression result is boolean
	boolVal, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("trigger did not return bool")
	}
	return boolVal, nil
}
