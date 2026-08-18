package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Recover retoma um run interrompido após restart do processo: adquire o
// lease limitado, aplica a política de recuperação do grafo e continua do
// último nó confirmado (tarefa 3.3). Versão de grafo não registrada é
// falha terminal estável, nunca migração silenciosa.
func (e *Engine) Recover(
	ctx context.Context,
	id RunID,
	owner string,
	lease time.Duration,
) (Run, error) {
	if e.leases == nil {
		return Run{}, errors.New("workflow: recuperação exige um LeaseStore configurado")
	}
	run, err := e.store.Get(ctx, id)
	if err != nil {
		return Run{}, err
	}
	if run.Status.Terminal() {
		return run, fmt.Errorf("workflow: run %s já tem desfecho terminal", id)
	}
	agora := e.clock.Now()
	obtido, err := e.leases.AcquireLease(ctx, id, owner, agora, agora.Add(lease))
	if err != nil {
		return run, err
	}
	if !obtido {
		return run, fmt.Errorf("%w: %s", ErrLeaseHeld, id)
	}
	defer func() { _ = e.leases.ReleaseLease(ctx, id, owner) }()
	def, err := e.registry.Get(run.Graph)
	if err != nil {
		return e.finish(ctx, run, RunFailed, OutcomeIncompatibleVersion, err)
	}
	switch def.Recovery {
	case RecoveryFail:
		return e.finish(ctx, run, RunFailed, OutcomeRecoveryRefused,
			fmt.Errorf("workflow: grafo %s não permite recuperação", run.Graph))
	case RecoveryRestart:
		run.Current = def.Initial
	}
	run.Status = RunActive
	run.StartedAt = agora // nova janela ativa após a interrupção
	run.UpdatedAt = agora
	run.Version++
	if err := e.store.Update(ctx, run); err != nil {
		return run, err
	}
	return e.loop(ctx, def, run)
}
