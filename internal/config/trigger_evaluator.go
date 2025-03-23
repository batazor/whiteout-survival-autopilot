package config

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// TriggerEvaluator interface
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
	if len(st.Accounts) == 0 || len(st.Accounts[0].Characters) == 0 {
		return false, fmt.Errorf("no character data available")
	}
	char := st.Accounts[0].Characters[0]

	// Flatten nested Gamer struct into map[string]interface{}
	flat := make(map[string]interface{})
	flattenStruct("", char, flat)

	// Create CEL env from keys dynamically
	var declsList []*exprpb.Decl
	for k, v := range flat {
		switch v.(type) {
		case bool:
			declsList = append(declsList, decls.NewVar(k, decls.Bool))
		case int, int64:
			declsList = append(declsList, decls.NewVar(k, decls.Int))
		case string:
			declsList = append(declsList, decls.NewVar(k, decls.String))
		}
	}

	env, err := cel.NewEnv(cel.Declarations(declsList...))
	if err != nil {
		return false, fmt.Errorf("creating CEL env: %w", err)
	}

	// Compile expression
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile error: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program creation error: %w", err)
	}

	// Eval with the flattened data
	out, _, err := prg.Eval(flat)
	if err != nil {
		return false, fmt.Errorf("eval error: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("trigger result is not bool: %v", out)
	}
	return result, nil
}
