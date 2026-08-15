package testkit

import (
	"context"
	"errors"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Run executa o cenário até desfecho terminal, aplicando as decisões de
// aprovação previstas a cada pausa; sem decisão prevista, o run permanece
// aguardando e o resultado reflete isso. O resultado captura também o
// checkpoint persistido e os efeitos pendentes para as asserções da 5.5.
func (h *Harness) Run(ctx context.Context) (Result, error) {
	run, err := h.inicia(ctx)
	for errors.Is(err, wf.ErrApprovalPending) {
		if !h.decidePendentes(ctx, run.ID) {
			break
		}
		run, err = h.Engine.Resume(ctx, run.ID)
	}
	steps, stepsErr := h.Store.Steps(ctx, run.ID)
	if stepsErr != nil {
		return Result{}, stepsErr
	}
	persistido, persistErr := h.Store.Get(ctx, run.ID)
	if persistErr != nil {
		persistido = wf.Run{} // crash antes do Create: sem checkpoint
	}
	pendentes, outboxErr := h.Outbox.Pending(ctx)
	if outboxErr != nil {
		return Result{}, outboxErr
	}
	return Result{
		Run:        run,
		Persistido: persistido,
		Steps:      steps,
		Events:     h.Events.Trilha(),
		Pendentes:  pendentes,
		Err:        err,
	}, nil
}

// Close libera os recursos do harness (o banco SQLite, quando usado).
func (h *Harness) Close() error {
	if fechavel, ok := h.Store.(interface{ Close() error }); ok {
		return fechavel.Close()
	}
	return nil
}

// decidePendentes aplica decisões previstas; false se nada foi decidido.
func (h *Harness) decidePendentes(ctx context.Context, id wf.RunID) bool {
	decidiu := false
	for _, ap := range h.Approvals.Pendentes(id) {
		dec, prevista := h.cenario.Aprovacoes[ap.Preview.Effect]
		if !prevista {
			continue
		}
		prazo := h.Clock.Now().Add(dec.Validade)
		if h.Approvals.Decide(ctx, id, ap.Hash, dec.Conceder, prazo) == nil {
			decidiu = true
		}
	}
	return decidiu
}
