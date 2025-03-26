package executor

import (
	"log/slog"
	"strings"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type UseCaseExecutor interface {
	ExecuteUseCase(uc *domain.UseCase, state *domain.State)
}

func NewUseCaseExecutor(logger *slog.Logger, triggerEvaluator config.TriggerEvaluator) UseCaseExecutor {
	return &executorImpl{
		logger:           logger,
		triggerEvaluator: triggerEvaluator,
	}
}

type executorImpl struct {
	logger           *slog.Logger
	triggerEvaluator config.TriggerEvaluator
}

func (e *executorImpl) ExecuteUseCase(uc *domain.UseCase, state *domain.State) {
	e.logger.Info("=== Start usecase ===", slog.String("name", uc.Name))
	for _, step := range uc.Steps {
		e.runStep(step, 0, state)
	}
	e.logger.Info("=== End usecase ===", slog.String("name", uc.Name))
}

func (e *executorImpl) runStep(step domain.Step, indent int, state *domain.State) bool {
	prefix := strings.Repeat("  ", indent)

	if step.Click != "" {
		e.logger.Info(prefix+"Click", slog.String("target", step.Click))
	}

	if step.Action != "" {
		e.logger.Info(prefix+"Action", slog.String("action", step.Action))

		switch step.Action {
		case "loop":
			if step.Trigger == "" {
				e.logger.Warn(prefix + "Loop trigger is missing, skipping loop")
				return false
			}
			e.logger.Info(prefix+"Entering loop", slog.String("trigger", step.Trigger))

			for {
				shouldContinue, err := e.triggerEvaluator.EvaluateTrigger(step.Trigger, state)
				if err != nil {
					e.logger.Error(prefix+"Trigger evaluation failed", slog.Any("error", err))
					break
				}
				if !shouldContinue {
					e.logger.Info(prefix + "Loop trigger returned false, exiting loop")
					break
				}

				for _, s := range step.Steps {
					stopped := e.runStep(s, indent+1, state)
					if stopped {
						e.logger.Info(prefix + "Loop stopped manually (loop_stop)")
						return false
					}
				}
			}

		case "loop_stop":
			e.logger.Info(prefix + "Received loop_stop")
			return true
		}
	}

	if step.Wait > 0 {
		e.logger.Info(prefix+"Wait", slog.Duration("duration", step.Wait))
		time.Sleep(step.Wait)
	}

	if step.If != nil {
		e.logger.Info(prefix+"If Trigger", slog.String("expr", step.If.Trigger))

		result, err := e.triggerEvaluator.EvaluateTrigger(step.If.Trigger, state)
		if err != nil {
			e.logger.Error(prefix+"Trigger evaluation failed",
				slog.String("expr", step.If.Trigger),
				slog.Any("error", err),
			)
			return false
		}

		if result {
			e.logger.Info(prefix + "Condition met: executing THEN")
			for _, s := range step.If.Then {
				stopped := e.runStep(s, indent+1, state)
				if stopped {
					return true
				}
			}
		} else if len(step.If.Else) > 0 {
			e.logger.Info(prefix + "Condition NOT met: executing ELSE")
			for _, s := range step.If.Else {
				stopped := e.runStep(s, indent+1, state)
				if stopped {
					return true
				}
			}
		}
	}

	return false
}
