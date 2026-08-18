package workflow

import (
	"context"
	"errors"
	"fmt"
)

// loop executa nós até desfecho terminal, limite, erro ou cancelamento.
// Cada iteração verifica, nesta ordem: cancelamento, limite de passos e
// deadline ativo — todos com outcomes estáveis (tarefa 2.4).
func (e *Engine) loop(ctx context.Context, def Definition, run Run) (Run, error) {
	deadline := run.StartedAt.Add(def.Limits.ActiveDeadline)
	ctx = withToolGuard(ctx, newToolGuard(def.Limits))
	if e.policy != nil {
		ctx = withEffectGate(ctx, &effectGate{
			policy: e.policy, approvals: e.approvals, clock: e.clock,
			run: run.ID, graph: run.Graph,
		})
	}
	for indice := 1; ; indice++ {
		if err := ctx.Err(); err != nil {
			return e.finish(ctx, run, RunCancelled, OutcomeCancelled, err)
		}
		if indice > def.Limits.MaxSteps {
			return e.finish(ctx, run, RunFailed, OutcomeStepLimit,
				fmt.Errorf("%w: passos > %d", ErrLimitExceeded, def.Limits.MaxSteps))
		}
		if e.clock.Now().After(deadline) {
			return e.finish(ctx, run, RunFailed, OutcomeDeadlineExceeded,
				fmt.Errorf("%w: deadline ativo", ErrLimitExceeded))
		}
		proximo, err := e.executaNo(ctx, def, &run, indice)
		if err != nil || proximo.EhTerminal() {
			return e.desfecho(ctx, run, proximo, err)
		}
		run.Current = proximo.Next
		run.Version++
		run.UpdatedAt = e.clock.Now()
		if err := e.store.Update(ctx, run); err != nil {
			return run, err
		}
	}
}

// executaNo roda o nó corrente, registra o step e resolve a transição.
func (e *Engine) executaNo(
	ctx context.Context,
	def Definition,
	run *Run,
	indice int,
) (Target, error) {
	no := def.Nodes[run.Current]
	inicio := e.clock.Now()
	estado, outcome, execErr := no.Execute(ctx, run.State)
	fim := e.clock.Now()
	if err := e.registraStep(ctx, *run, indice, outcome, execErr, inicio, fim); err != nil {
		return Target{}, err
	}
	if execErr != nil {
		return Target{}, execErr
	}
	if err := validaPayload(estado, def.Limits.MaxPayloadBytes); err != nil {
		return Target{}, err
	}
	run.State = estado
	alvo, ok := def.Transitions[TransitionKey{From: run.Current, Outcome: outcome}]
	if !ok {
		return Target{}, fmt.Errorf("%w: %q em %q", ErrUnmappedOutcome, outcome, run.Current)
	}
	return alvo, nil
}

// desfecho traduz o resultado de executaNo em status e outcome estáveis.
func (e *Engine) desfecho(ctx context.Context, run Run, alvo Target, err error) (Run, error) {
	switch {
	case err == nil:
		status := RunSucceeded
		if alvo.Falha {
			status = RunFailed
		}
		return e.finish(ctx, run, status, alvo.Terminal, nil)
	case errors.Is(err, ErrApprovalPending):
		return e.finish(ctx, run, RunWaiting, "", err)
	case errors.Is(err, ErrPolicyDenied):
		return e.finish(ctx, run, RunFailed, OutcomePolicyDenied, err)
	case errors.Is(err, ErrApprovalExpired):
		return e.finish(ctx, run, RunFailed, OutcomeApprovalExpired, err)
	case errors.Is(err, ErrUnmappedOutcome):
		return e.finish(ctx, run, RunFailed, OutcomeUnmapped, err)
	case errors.Is(err, ErrPayloadExceeded):
		return e.finish(ctx, run, RunFailed, OutcomePayloadLimit, err)
	case errors.Is(err, ErrLimitExceeded):
		return e.finish(ctx, run, RunFailed, OutcomeToolLimit, err)
	default:
		return e.finish(ctx, run, RunFailed, OutcomeNodeError, err)
	}
}
