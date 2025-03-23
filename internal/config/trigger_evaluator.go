package config

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type TriggerEvaluator interface {
	EvaluateTrigger(expr string, state map[string]interface{}) (bool, error)
}

func NewTriggerEvaluator() TriggerEvaluator {
	return &triggerEvaluator{}
}

type triggerEvaluator struct{}

func (t *triggerEvaluator) EvaluateTrigger(expr string, state map[string]interface{}) (bool, error) {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("state", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return false, fmt.Errorf("creating CEL env: %w", err)
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile error: %v", issues.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("env.Program error: %w", err)
	}

	out, _, err := prg.Eval(map[string]interface{}{"state": state})
	if err != nil {
		return false, fmt.Errorf("eval error: %w", err)
	}
	boolVal, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("expression did not return bool")
	}
	return boolVal, nil
}
