package domain

import (
	"fmt"
	"time"
)

// UseCase представляет сценарий, описанный в YAML-файле.
type UseCase struct {
	Name     string        `yaml:"name"`     // Название сценария
	Priority int           `yaml:"priority"` // Приоритет сценария (от 0 до 100)
	Node     string        `yaml:"node"`     // Начальный экран/состояние, с которого начинается usecase
	Trigger  string        `yaml:"trigger"`  // CEL-выражение, которое определяет, запускать ли usecase
	Steps    Steps         `yaml:"steps"`    // Последовательность шагов
	TTL      time.Duration `yaml:"ttl"`      // Время жизни usecase (например, "24h")
	Cron     string        `yaml:"cron"`     // Cron-выражение для периодического запуска usecase (например, "0 0 * * *")

	SourcePath string `json:"-"` // Путь к файлу, из которого был загружен usecase
}

// Steps — это просто срез Step
type Steps []Step

// Step представляет отдельный шаг в usecase.
// Он может быть простым действием (click/wait), условным (if), циклическим (loop),
// включать анализ скриншота (analyze), или управлять TTL.
type Step struct {
	// Общие действия
	Click   string        `yaml:"click,omitempty"`   // Название региона, по которому нужно кликнуть
	Longtap string        `yaml:"longtap,omitempty"` // Название региона, по которому нужно сделать долгий тап
	Action  string        `yaml:"action,omitempty"`  // Специальное действие: "loop", "loop_stop", "screenshot", и т.д.
	Wait    time.Duration `yaml:"wait,omitempty"`    // Ожидание (например, "5s")

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

	Set string      `yaml:"set,omitempty"` // ← сюда попадёт path вроде "exploration.state.battleStatus"
	To  interface{} `yaml:"to,omitempty"`  // ← новое значение (в твоём случае — "")
}

// IfStep описывает условную конструкцию вида if { then {} else {} }
type IfStep struct {
	Trigger string `yaml:"trigger"`        // CEL-выражение, возвращающее bool
	Then    []Step `yaml:"then"`           // Шаги, выполняемые, если trigger = true
	Else    []Step `yaml:"else,omitempty"` // Шаги, если trigger = false (опционально)
}

// AnalyzeRule описывает правила для анализа региона экрана (screenshot).
type AnalyzeRule struct {
	Name          string            `yaml:"name"`                    // Название региона (и ключ для сохранения)
	Action        string            `yaml:"action"`                  // Действие: "text", "exist", "color_check", "findIcon", "findText"
	Text          string            `yaml:"text,omitempty"`          // Текст для поиска (например, "Battle")
	Type          string            `yaml:"type,omitempty"`          // Тип результата (например, "integer", если action = text)
	Threshold     float64           `yaml:"threshold,omitempty"`     // Уровень уверенности, по умолчанию 0.9
	ExpectedColor string            `yaml:"expectedColor,omitempty"` // Цвет для проверки (например, "green")
	Log           string            `yaml:"log,omitempty"`           // Сообщение для логирования (опционально)
	SaveAsRegion  bool              `yaml:"saveAsRegion,omitempty"`  // если true — сохранить зону как новую временную область с именем .Name
	Options       *AnalyzeImageRule `yaml:"options,omitempty"`       // Опции для анализа изображения
	PushUseCase   []PushUsecase     `yaml:"pushUsecase,omitempty"`   // Список usecase, которые нужно запустить при выполнении этого правила
}

type PushUsecase struct {
	Trigger string    `yaml:"trigger"` // CEL-выражение
	List    []UseCase `yaml:"list"`    // Юзкейсы, которые нужно отправить в очередь
}

// Validate проверяет допустимость значения action в правиле анализа.
func (r AnalyzeRule) Validate() error {
	switch r.Action {
	case "text", "exist", "color_check", "findIcon", "findText":
		return nil
	default:
		return fmt.Errorf("invalid action '%s' in rule '%s'", r.Action, r.Name)
	}
}

type AnalyzeImageRule struct {
	Clane bool `yaml:"clane,omitempty"` // если true — использовать CLAHE
}
