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
		step.Err = codigoEstavel(execErr)
	}
	if err := e.store.AppendStep(ctx, step); err != nil {
		return fmt.Errorf("workflow: falha ao registrar step: %w", err)
	}
	if e.events == nil {
		return nil
	}
	ev := Event{
		Schema: EventSchema, Run: run.ID, Step: indice, Graph: run.Graph,
		Correlation: string(run.ID), At: fim, Outcome: outcome,
		DurationMS: fim.Sub(inicio).Milliseconds(),
	}
	if indice > 1 {
		ev.Causation = fmt.Sprintf("step-%d", indice-1)
	}
	if step.Err != "" {
		ev.Attrs = map[string]string{"error_code": step.Err}
	}
	return e.events.Emit(ctx, ev)
}

// codigoEstavel traduz o erro tipado da tentativa em código estável de
// evidência; o texto da mensagem nunca vai para o journal.
func codigoEstavel(execErr error) string {
	switch {
	case errors.Is(execErr, ErrApprovalPending):
		return "approval_pending"
	case errors.Is(execErr, ErrPolicyDenied):
		return string(OutcomePolicyDenied)
	case errors.Is(execErr, ErrApprovalExpired):
		return string(OutcomeApprovalExpired)
	case errors.Is(execErr, ErrPayloadExceeded):
		return string(OutcomePayloadLimit)
	case errors.Is(execErr, ErrLimitExceeded):
		return string(OutcomeToolLimit)
	default:
		return string(OutcomeNodeError)
	}
}
