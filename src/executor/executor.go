package executor

import (
	"fmt"
	"log"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/src/config"
)

// ExecuteUseCase симулирует выполнение сценария, проходя по всем шагам.
func ExecuteUseCase(uc *config.UseCase) {
	log.Printf("Начало выполнения usecase: %s", uc.Name)
	for _, step := range uc.Steps {
		executeStep(step, 0)
	}
	log.Printf("Завершено выполнение usecase: %s", uc.Name)
}

// executeStep рекурсивно выполняет шаг и выводит информацию в консоль.
func executeStep(step config.Step, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	if step.Click != "" {
		fmt.Printf("%sКлик: %s\n", prefix, step.Click)
	}
	if step.Action != "" {
		fmt.Printf("%sДействие: %s\n", prefix, step.Action)
	}
	if step.Wait != 0 {
		dur := time.Duration(step.Wait)
		fmt.Printf("%sОжидание: %s\n", prefix, dur.String())
		time.Sleep(dur)
	}
	if step.If != nil {
		fmt.Printf("%sУсловие: %s\n", prefix, step.If.Trigger)
		// Для демонстрации просто выбираем ветку Then.
		// В реальном приложении условие необходимо оценивать через EvaluateTrigger.
		fmt.Printf("%sВыполнение ветки Then:\n", prefix)
		for _, s := range step.If.Then {
			executeStep(s, indent+1)
		}
		if len(step.If.Else) > 0 {
			fmt.Printf("%sВыполнение ветки Else:\n", prefix)
			for _, s := range step.If.Else {
				executeStep(s, indent+1)
			}
		}
	}
}
