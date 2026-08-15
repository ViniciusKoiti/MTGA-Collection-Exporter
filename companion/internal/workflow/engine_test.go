package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// ambiente monta engine + adapters in-memory com um grafo linear de dois nós.
func ambiente(t *testing.T, nos map[wf.NodeID]wf.Node, limites wf.Limits) (*wf.Engine, *memory.Clock, *memory.RunStore) {
	t.Helper()
	registry := wf.NewRegistry(wf.Limits{MaxSteps: 100, ActiveDeadline: time.Hour})
	def := wf.Definition{
		Identity: wf.Identity{Kind: "collection-sync", Version: 1},
		Initial:  "coleta",
		Nodes:    nos,
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "coleta", Outcome: "ok"}:  {Next: "valida"},
			{From: "valida", Outcome: "ok"}:  {Terminal: "done"},
			{From: "valida", Outcome: "vai"}: {Next: "coleta"},
		},
		Limits:   limites,
		Recovery: wf.RecoveryResume,
	}
	if err := registry.Register(def); err != nil {
		t.Fatalf("registro: %v", err)
	}
	clock := memory.NewClock(time.Unix(1_700_000_000, 0))
	store := memory.NewRunStore()
	return wf.NewEngine(registry, clock, memory.NewIDs("run"), store), clock, store
}

type noFn struct {
	id wf.NodeID
	fn func(context.Context, wf.State) (wf.State, wf.OutcomeCode, error)
}

func (n noFn) ID() wf.NodeID { return n.id }
func (n noFn) Execute(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
	return n.fn(ctx, st)
}

func passa(id wf.NodeID, out wf.OutcomeCode) wf.Node {
	return noFn{id, func(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		return st, out, nil
	}}
}

func TestRunLinearSucedeComEvidencia(t *testing.T) {
	engine, _, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	run, err := engine.Start(t.Context(), wf.Identity{Kind: "collection-sync", Version: 1}, 7)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if run.Status != wf.RunSucceeded || run.Outcome != "done" {
		t.Fatalf("desfecho inesperado: %+v", run)
	}
	persistido, err := store.Get(t.Context(), run.ID)
	if err != nil || persistido.Status != wf.RunSucceeded {
		t.Fatalf("checkpoint final não persistido: %+v (%v)", persistido, err)
	}
	steps, _ := store.Steps(t.Context(), run.ID)
	if len(steps) != 2 || steps[0].Node != "coleta" || steps[1].Node != "valida" {
		t.Fatalf("journal de steps inesperado: %+v", steps)
	}
}

func TestOutcomeSemTransicaoFalhaEstavel(t *testing.T) {
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "inventado"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	run, err := engine.Start(t.Context(), wf.Identity{Kind: "collection-sync", Version: 1}, nil)
	if !errors.Is(err, wf.ErrUnmappedOutcome) {
		t.Fatalf("esperava ErrUnmappedOutcome, veio: %v", err)
	}
	if run.Status != wf.RunFailed || run.Outcome != wf.OutcomeUnmapped {
		t.Fatalf("desfecho inesperado: %+v", run)
	}
}
