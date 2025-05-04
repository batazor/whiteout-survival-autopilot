package config

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/ext"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

type TriggerEvaluator interface {
	EvaluateTrigger(expr string, st *domain.Gamer) (bool, error)
}

func NewTriggerEvaluator() TriggerEvaluator {
	return &triggerEvaluator{}
}

// -----------------------------------------------------------------------------
// Implementation
// -----------------------------------------------------------------------------

type triggerEvaluator struct{}

// EvaluateTrigger compiles the CEL expression (expr) and evaluates it against
// the flattened *domain.Gamer.
func (t *triggerEvaluator) EvaluateTrigger(expr string, char *domain.Gamer) (bool, error) {
	// 1. Flatten struct -> map[string]interface{}
	flat := make(map[string]interface{})
	flattenStruct("", char, flat)

	// 2. Declarations for every known field
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

	// 3. CEL environment
	env, err := cel.NewEnv(
		cel.Declarations(declsList...),
		ext.Strings(), // adds string.lowerAscii()
		CompareTextLib,
	)
	if err != nil {
		return false, fmt.Errorf("creating CEL env: %w", err)
	}

	// 4. Compile & run
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile error: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program creation error: %w", err)
	}

	out, _, err := prg.Eval(flat)
	if err != nil {
		return false, fmt.Errorf("eval error: %w", err)
	}

	// 5. Expecting bool
	res, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("trigger result is not bool: %v", out)
	}

	return res, nil
}
