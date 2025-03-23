package config

import (
	"context"
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/spf13/viper"
)

// UseCase представляет сценарий, описанный в YAML-файле.
type UseCase struct {
	Name      string `yaml:"name"`       // Название сценария
	Node      string `yaml:"node"`       // Начальное состояние (экран)
	Trigger   string `yaml:"trigger"`    // Условие запуска сценария (CEL-выражение)
	Steps     []Step `yaml:"steps"`      // Последовательность шагов сценария
	FinalNode string `yaml:"final_node"` // Конечное состояние после выполнения сценария
}

// Step представляет отдельный шаг сценария.
// Поддерживаются действия клика, выполнение дополнительных действий, ожидание и условные операторы.
type Step struct {
	SetTTL      string `yaml:"setTTL,omitempty"`      // TTL duration (e.g. "24h") to set in Redis
	UsecaseName string `yaml:"usecaseName,omitempty"` // Target usecase name for the TTL

	Click  string        `yaml:"click,omitempty"`  // Клик по элементу (например, "to_mail")
	Action string        `yaml:"action,omitempty"` // Дополнительное действие (например, "screenshot")
	Wait   time.Duration `yaml:"wait,omitempty"`   // Ожидание, заданное строкой, например, "5s"
	If     *IfStep       `yaml:"if,omitempty"`     // Условный оператор, если шаг включает проверку
}

// IfStep описывает условный шаг с ветками then/else.
// Trigger — условие, которое будет оцениваться с помощью cel-go.
type IfStep struct {
	Trigger string `yaml:"trigger"`        // Условие (CEL-выражение)
	Then    []Step `yaml:"then"`           // Шаги, выполняемые при истинном условии
	Else    []Step `yaml:"else,omitempty"` // Шаги, выполняемые при ложном условии (опционально)
}

// LoadUseCase загружает YAML-сценарий из файла с использованием Viper.
func LoadUseCase(ctx context.Context, configFile string) (*UseCase, error) {
	v := viper.New()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var uc UseCase
	if err := v.Unmarshal(&uc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal usecase: %w", err)
	}

	// Дополнительная валидация может быть добавлена при необходимости.
	return &uc, nil
}

// EvaluateTrigger использует cel-go для оценки CEL-выражения.
// trigger - CEL-выражение, state - карта текущего состояния (например, {"isNewMessage": true}).
// Функция возвращает результат выражения (bool) или ошибку.
func EvaluateTrigger(trigger string, state map[string]interface{}) (bool, error) {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("state", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return false, fmt.Errorf("failed to create CEL env: %v", err)
	}

	ast, issues := env.Compile(trigger)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("failed to compile trigger expression: %v", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("failed to create CEL program: %v", err)
	}

	result, _, err := prg.Eval(map[string]interface{}{"state": state})
	if err != nil {
		return false, fmt.Errorf("failed to evaluate trigger: %v", err)
	}

	if boolResult, ok := result.Value().(bool); ok {
		return boolResult, nil
	}

	return false, fmt.Errorf("trigger expression did not return a bool")
}
