package domain

import (
	"time"
)

// UseCase представляет сценарий, описанный в YAML-файле.
type UseCase struct {
	Name      string `yaml:"name"`                                  // Название сценария
	Node      string `yaml:"node"`                                  // Начальное состояние (экран)
	Trigger   string `yaml:"trigger"`                               // Условие запуска сценария (CEL-выражение)
	Steps     []Step `yaml:"steps"`                                 // Последовательность шагов сценария
	FinalNode string `yaml:"final_node"  mapstructure:"final_node"` // Конечное состояние после выполнения сценария
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
