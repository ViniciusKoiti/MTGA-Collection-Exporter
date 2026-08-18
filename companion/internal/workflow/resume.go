package workflow

import (
	"context"
	"fmt"
)

// Resume retoma um run pausado em waiting_approval a partir do último
// checkpoint. O nó corrente é reexecutado; com a aprovação decidida, o
// effect gate libera (concedida), nega (recusada) ou expira o efeito.
// A espera não consome deadline ativo: a retomada abre nova janela.
func (e *Engine) Resume(ctx context.Context, id RunID) (Run, error) {
	run, err := e.store.Get(ctx, id)
	if err != nil {
		return Run{}, err
	}
	if run.Status != RunWaiting {
		return run, fmt.Errorf("workflow: run %s não está aguardando aprovação (status %s)",
			id, run.Status)
	}
	def, err := e.registry.Get(run.Graph)
	if err != nil {
		return run, err
	}
	agora := e.clock.Now()
	run.Status = RunActive
	run.StartedAt = agora // nova janela ativa (espera não consome deadline)
	run.UpdatedAt = agora
	run.Version++
	if err := e.store.Update(ctx, run); err != nil {
		return run, err
	}
	return e.loop(ctx, def, run)
}
