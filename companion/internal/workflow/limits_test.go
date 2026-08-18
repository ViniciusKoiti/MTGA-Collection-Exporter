package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func identidade() wf.Identity { return wf.Identity{Kind: "collection-sync", Version: 1} }

func TestLimiteDePassosEncerraCicloComOutcomeEstavel(t *testing.T) {
	// valida devolve "vai" e volta para coleta: ciclo infinito sem limite.
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "vai"),
	}, wf.Limits{MaxSteps: 5})
	run, err := engine.Start(t.Context(), identidade(), nil)
	if !errors.Is(err, wf.ErrLimitExceeded) {
		t.Fatalf("esperava ErrLimitExceeded, veio: %v", err)
	}
	if run.Status != wf.RunFailed || run.Outcome != wf.OutcomeStepLimit {
		t.Fatalf("desfecho inesperado: %+v", run)
	}
}

func TestDeadlineAtivoUsaRelogioControlado(t *testing.T) {
	var clock interface{ Advance(time.Duration) }
	lento := noFn{"coleta", func(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		clock.Advance(time.Minute) // o nó consome mais que o deadline
		return st, "ok", nil
	}}
	engine, c, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": lento, "valida": passa("valida", "ok"),
	}, wf.Limits{ActiveDeadline: 30 * time.Second})
	clock = c
	run, err := engine.Start(t.Context(), identidade(), nil)
	if !errors.Is(err, wf.ErrLimitExceeded) || run.Outcome != wf.OutcomeDeadlineExceeded {
		t.Fatalf("esperava deadline excedido, veio: %v / %+v", err, run)
	}
}

func TestPayloadAcimaDoLimiteFalhaEstavel(t *testing.T) {
	gordo := noFn{"coleta", func(_ context.Context, _ wf.State) (wf.State, wf.OutcomeCode, error) {
		return string(make([]byte, 4096)), "ok", nil
	}}
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": gordo, "valida": passa("valida", "ok"),
	}, wf.Limits{MaxPayloadBytes: 64})
	run, err := engine.Start(t.Context(), identidade(), nil)
	if !errors.Is(err, wf.ErrLimitExceeded) || run.Outcome != wf.OutcomePayloadLimit {
		t.Fatalf("esperava limite de payload, veio: %v / %+v", err, run)
	}
}

func TestCancelamentoDoContextoEncerraRun(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancelador := noFn{"coleta", func(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		cancel()
		return st, "ok", nil
	}}
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": cancelador, "valida": passa("valida", "ok"),
	}, wf.Limits{})
	run, err := engine.Start(ctx, identidade(), nil)
	if !errors.Is(err, context.Canceled) || run.Outcome != wf.OutcomeCancelled {
		t.Fatalf("esperava cancelamento, veio: %v / %+v", err, run)
	}
}
