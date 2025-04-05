package config

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/ext"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type TriggerEvaluator interface {
	EvaluateTrigger(expr string, st *domain.Gamer) (bool, error)
}

func NewTriggerEvaluator() TriggerEvaluator {
	return &triggerEvaluator{}
}

type triggerEvaluator struct{}

// EvaluateTrigger compiles the CEL expression (expr) and then evaluates it
// against the data we extract from *domain.Gamer.
func (t *triggerEvaluator) EvaluateTrigger(expr string, char *domain.Gamer) (bool, error) {
	// Flatten nested Gamer struct into map[string]interface{}
	flat := make(map[string]interface{})
	flattenStruct("", char, flat)

	// Build a list of declarations for every known field
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

	// Создаём окружение CEL с нашими переменными + библиотекой ext.Strings()
	// ext.Strings() автоматически добавляет метод: string.lowerAscii()
	env, err := cel.NewEnv(
		cel.Declarations(declsList...),
		ext.Strings(),
	)
	if err != nil {
		return false, fmt.Errorf("creating CEL env: %w", err)
	}

	// Компилируем выражение
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile error: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program creation error: %w", err)
	}

	// Выполняем
	out, _, err := prg.Eval(flat)
	if err != nil {
		return false, fmt.Errorf("eval error: %w", err)
	}

	// Ожидаем bool
	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("trigger result is not bool: %v", out)
	}
	return result, nil
}
