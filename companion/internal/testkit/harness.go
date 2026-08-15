// Package testkit é o harness de desenvolvimento determinístico do OpenSpec
// add-graph-workflow-harness (decisão 5 do design): compõe o engine DE
// PRODUÇÃO com banco temporário (in-memory ou SQLite migrado), relógio e
// IDs controlados, decisões de aprovação previstas, decoradores de falha e
// entry points de atividade. Ele nunca reimplementa transições.
package testkit

import (
	"context"
	"errors"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/sqlitestore"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// Harness expõe o engine composto e os adapters controlados do cenário.
type Harness struct {
	Engine    *wf.Engine
	Clock     *memory.Clock
	Store     wf.RunStore
	Approvals *memory.Approvals
	Outbox    *memory.Outbox
	Events    *memory.Events
	launcher  *activity.Launcher
	cenario   Scenario
}

// New compõe o perfil de produção do engine com os adapters do harness.
func New(sc Scenario) (*Harness, error) {
	h := &Harness{
		Clock:     memory.NewClock(sc.Inicio),
		Approvals: memory.NewApprovals(),
		Outbox:    memory.NewOutbox(),
		Events:    memory.NewEvents(),
		cenario:   sc,
	}
	registry := wf.NewRegistry(wf.Limits{})
	if err := sc.Registrar(registry, h.Clock); err != nil {
		return nil, err
	}
	if sc.CaminhoBanco != "" {
		banco, err := sqlitestore.Open(sc.CaminhoBanco) // migrações de produção
		if err != nil {
			return nil, err
		}
		h.Store = banco
	} else {
		h.Store = memory.NewRunStore()
	}
	store := h.Store
	if sc.DecorarStore != nil {
		store = sc.DecorarStore(store)
	}
	h.Engine = wf.NewEngine(registry, h.Clock, memory.NewIDs("run"), store)
	h.Engine.WithEvents(h.Events)
	if sc.Policy != nil {
		h.Engine.WithPolicy(sc.Policy, h.Approvals)
	}
	if sc.Atividade != "" {
		if sc.Inventario == nil {
			return nil, errors.New("testkit: cenário com atividade exige inventário")
		}
		launcher, err := activity.NewLauncher(sc.Inventario, registry, h.Engine)
		if err != nil {
			return nil, err
		}
		h.launcher = launcher
	}
	return h, nil
}

// inicia executa pelo entry point de produção quando o cenário nomeia uma
// atividade; caso contrário chama o engine diretamente.
func (h *Harness) inicia(ctx context.Context) (wf.Run, error) {
	if h.launcher != nil {
		return h.launcher.Run(ctx, h.cenario.Atividade, h.cenario.Estado)
	}
	return h.Engine.Start(ctx, h.cenario.Graph, h.cenario.Estado)
}
