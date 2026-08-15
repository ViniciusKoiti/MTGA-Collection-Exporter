// Package testkit é o harness de desenvolvimento determinístico do OpenSpec
// add-graph-workflow-harness (decisão 5 do design): compõe o engine DE
// PRODUÇÃO com banco temporário em memória, relógio e IDs controlados,
// decisões de aprovação previstas e decoradores de falha. Ele nunca
// reimplementa transições.
package testkit

import (
	"context"
	"errors"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// Harness expõe o engine composto e os adapters controlados do cenário.
type Harness struct {
	Engine    *wf.Engine
	Clock     *memory.Clock
	Store     *memory.RunStore
	Approvals *memory.Approvals
	Outbox    *memory.Outbox
	Events    *memory.Events
	cenario   Scenario
}

// New compõe o perfil de produção do engine com os adapters do harness.
func New(sc Scenario) (*Harness, error) {
	h := &Harness{
		Clock:     memory.NewClock(sc.Inicio),
		Store:     memory.NewRunStore(),
		Approvals: memory.NewApprovals(),
		Outbox:    memory.NewOutbox(),
		Events:    memory.NewEvents(),
		cenario:   sc,
	}
	registry := wf.NewRegistry(wf.Limits{})
	if err := sc.Registrar(registry, h.Clock); err != nil {
		return nil, err
	}
	var store wf.RunStore = h.Store
	if sc.DecorarStore != nil {
		store = sc.DecorarStore(store)
	}
	h.Engine = wf.NewEngine(registry, h.Clock, memory.NewIDs("run"), store)
	h.Engine.WithEvents(h.Events)
	if sc.Policy != nil {
		h.Engine.WithPolicy(sc.Policy, h.Approvals)
	}
	return h, nil
}

// Run executa o cenário até desfecho terminal, aplicando as decisões de
// aprovação previstas a cada pausa; sem decisão prevista, o run permanece
// aguardando e o resultado reflete isso.
func (h *Harness) Run(ctx context.Context) (Result, error) {
	run, err := h.Engine.Start(ctx, h.cenario.Graph, h.cenario.Estado)
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
	return Result{Run: run, Steps: steps, Events: h.Events.Trilha(), Err: err}, nil
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
