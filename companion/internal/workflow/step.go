package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// registraStep grava a evidência da tentativa sem payloads (spec
// workflow-observability: nada de estado, entradas ou saídas no journal).
func (e *Engine) registraStep(
	ctx context.Context,
	run Run,
	indice int,
	outcome OutcomeCode,
	execErr error,
	inicio, fim time.Time,
) error {
	step := Step{
		Run:      run.ID,
		Index:    indice,
		Node:     run.Current,
		Attempt:  1,
		Outcome:  outcome,
		Started:  inicio,
		Finished: fim,
	}
	if execErr != nil {
		step.Err = string(OutcomeNodeError)
		if errors.Is(execErr, ErrLimitExceeded) {
			step.Err = string(OutcomeToolLimit)
		}
	}
	if err := e.store.AppendStep(ctx, step); err != nil {
		return fmt.Errorf("workflow: falha ao registrar step: %w", err)
	}
	return nil
}
