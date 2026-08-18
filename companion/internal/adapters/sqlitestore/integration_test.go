package sqlitestore

// Helpers dos testes de integração da tarefa 3.7: engine de produção
// composto sobre o Store SQLite real.

import (
	"context"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

type noSimples struct {
	id  wf.NodeID
	out wf.OutcomeCode
}

func (n noSimples) ID() wf.NodeID { return n.id }
func (n noSimples) Execute(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
	return st, n.out, nil
}

func registryLinear(t *testing.T) *wf.Registry {
	t.Helper()
	registry := wf.NewRegistry(wf.Limits{})
	err := registry.Register(wf.Definition{
		Identity: wf.Identity{Kind: "collection-sync", Version: 1},
		Initial:  "coleta",
		Nodes: map[wf.NodeID]wf.Node{
			"coleta": noSimples{"coleta", "ok"},
			"valida": noSimples{"valida", "ok"},
		},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "coleta", Outcome: "ok"}: {Next: "valida"},
			{From: "valida", Outcome: "ok"}: {Terminal: "done"},
		},
		Recovery: wf.RecoveryResume,
	})
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	return registry
}

// engineSobre compõe um engine de produção sobre o RunStore dado, com o
// Store SQLite como LeaseStore.
func engineSobre(t *testing.T, store wf.RunStore, leases wf.LeaseStore, prefixo string) (*wf.Engine, *memory.Clock) {
	t.Helper()
	clock := memory.NewClock(time.Unix(1_700_000_000, 0).UTC())
	engine := wf.NewEngine(registryLinear(t), clock, memory.NewIDs(prefixo), store)
	engine.WithRecovery(leases)
	return engine, clock
}

// interrompido grava um run ativo no meio do grafo, como se o processo
// tivesse morrido após confirmar o primeiro nó.
func interrompido(t *testing.T, store *Store, id wf.RunID) {
	t.Helper()
	em := time.Unix(1_700_000_000, 0).UTC()
	run := wf.Run{ID: id, Graph: wf.Identity{Kind: "collection-sync", Version: 1},
		Status: wf.RunActive, Current: "valida", Version: 1, StartedAt: em, UpdatedAt: em}
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
}
