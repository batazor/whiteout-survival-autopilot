package domain

import (
	"time"
)

// UseCase представляет сценарий, описанный в YAML-файле.
type UseCase struct {
	Name    string `yaml:"name"`    // Название сценария
	Node    string `yaml:"node"`    // Начальный экран/состояние, с которого начинается usecase
	Trigger string `yaml:"trigger"` // CEL-выражение, которое определяет, запускать ли usecase
	Steps   Steps  `yaml:"steps"`   // Последовательность шагов
}

// Steps — это просто срез Step
type Steps []Step

// Step представляет отдельный шаг в usecase.
// Он может быть простым действием (click/wait), условным (if), циклическим (loop),
// включать анализ скриншота (analyze), или управлять TTL.
type Step struct {
	// Общие действия
	Click  string        `yaml:"click,omitempty"`  // Название региона, по которому нужно кликнуть
	Action string        `yaml:"action,omitempty"` // Специальное действие: "loop", "loop_stop", "screenshot", и т.д.
	Wait   time.Duration `yaml:"wait,omitempty"`   // Ожидание (например, "5s")

	// Условный блок if { then {} else {} }
	If *IfStep `yaml:"if,omitempty"`

	// Цикл: используется вместе с action: loop
	Trigger string `yaml:"trigger,omitempty"` // CEL-выражение, используемое для loop или if
	Steps   Steps  `yaml:"steps,omitempty"`   // Вложенные шаги (например, внутри loop или if.then)

	// Анализ скриншота: используется в паре с action: screenshot
	Analyze []AnalyzeRule `yaml:"analyze,omitempty"` // Список правил анализа (например, text/icon/etc.)

	// Управление TTL (например, чтобы отложить повторное выполнение usecase)
	SetTTL      string `yaml:"setTTL,omitempty"`      // Продолжительность, например "24h"
	UsecaseName string `yaml:"usecaseName,omitempty"` // Название usecase, к которому применяется TTL
}

// IfStep описывает условную конструкцию вида if { then {} else {} }
type IfStep struct {
	Trigger string `yaml:"trigger"`        // CEL-выражение, возвращающее bool
	Then    []Step `yaml:"then"`           // Шаги, выполняемые, если trigger = true
	Else    []Step `yaml:"else,omitempty"` // Шаги, если trigger = false (опционально)
}

// AnalyzeRule описывает правила для анализа региона экрана (screenshot).
type AnalyzeRule struct {
	Name          string  `yaml:"name"`                     // Название региона (и ключ для сохранения)
	Action        string  `yaml:"action"`                   // Действие: "text", "exist", "color_check"
	Type          string  `yaml:"type,omitempty"`           // Тип результата (например, "integer", если action = text)
	Threshold     float64 `yaml:"threshold,omitempty"`      // Уровень уверенности, по умолчанию 0.9
	ExpectedColor string  `yaml:"expected_color,omitempty"` // Цвет для проверки (например, "green")
	Set           string  `yaml:"set,omitempty"`            // Путь к полю, куда сохранить результат
	Log           string  `yaml:"log,omitempty"`            // Сообщение для логирования (опционально)
}
