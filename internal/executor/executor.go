package executor

import (
	"fmt"
	"log"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type UseCaseExecutor interface {
	ExecuteUseCase(uc *domain.UseCase)
}

func NewUseCaseExecutor() UseCaseExecutor {
	return &executorImpl{}
}

type executorImpl struct{}

func (e *executorImpl) ExecuteUseCase(uc *domain.UseCase) {
	log.Printf("=== Start usecase: %s ===", uc.Name)
	for _, step := range uc.Steps {
		e.runStep(step, 0)
	}
	log.Printf("=== End usecase: %s ===", uc.Name)
}

// runStep is recursive if you have nested If -> Then/Else steps
func (e *executorImpl) runStep(step domain.Step, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	if step.Click != "" {
		fmt.Printf("%sClick: %s\n", prefix, step.Click)
		// Here you'd do the actual click action...
	}
	if step.Action != "" {
		fmt.Printf("%sAction: %s\n", prefix, step.Action)
	}
	if step.Wait > 0 {
		fmt.Printf("%sWaiting for: %v...\n", prefix, step.Wait)
		time.Sleep(step.Wait)
	}
	if step.If != nil {
		fmt.Printf("%sIF: %s\n", prefix, step.If.Trigger)
		// Evaluate condition in real code. For now, we pretend it’s true
		conditionTrue := true

		if conditionTrue {
			fmt.Printf("%s  THEN:\n", prefix)
			for _, s := range step.If.Then {
				e.runStep(s, indent+2)
			}
		} else if len(step.If.Else) > 0 {
			fmt.Printf("%s  ELSE:\n", prefix)
			for _, s := range step.If.Else {
				e.runStep(s, indent+2)
			}
		}
	}
}
