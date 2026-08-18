package workflow_test

import (
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// runInterrompido simula processo morto após confirmar o nó "valida".
func runInterrompido(t *testing.T, store *memory.RunStore, versao int) wf.Run {
	t.Helper()
	run := wf.Run{
		ID: "run-recuperavel", Graph: wf.Identity{Kind: "collection-sync", Version: versao},
		Status: wf.RunActive, Current: "valida", Version: 1,
		StartedAt: time.Unix(1_700_000_000, 0), UpdatedAt: time.Unix(1_700_000_000, 0),
	}
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
	return run
}

func TestRecoverRetomaDoUltimoNoConfirmado(t *testing.T) {
	engine, _, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	engine.WithRecovery(store)
	run := runInterrompido(t, store, 1)
	recuperado, err := engine.Recover(t.Context(), run.ID, "worker-1", time.Minute)
	if err != nil || recuperado.Status != wf.RunSucceeded || recuperado.Outcome != "done" {
		t.Fatalf("recuperação deveria concluir do nó confirmado: %v / %+v", err, recuperado)
	}
	steps, _ := store.Steps(t.Context(), run.ID)
	if len(steps) != 1 || steps[0].Node != "valida" {
		t.Fatalf("apenas o nó pendente deveria reexecutar: %+v", steps)
	}
}

func TestRecoverRespeitaLeaseDeOutroDono(t *testing.T) {
	engine, clock, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	engine.WithRecovery(store)
	run := runInterrompido(t, store, 1)
	agora := clock.Now()
	if ok, err := store.AcquireLease(t.Context(), run.ID, "worker-vivo", agora, agora.Add(time.Hour)); !ok || err != nil {
		t.Fatalf("lease inicial: %v", err)
	}
	if _, err := engine.Recover(t.Context(), run.ID, "worker-2", time.Minute); !errors.Is(err, wf.ErrLeaseHeld) {
		t.Fatalf("esperava ErrLeaseHeld, veio: %v", err)
	}
	clock.Advance(2 * time.Hour) // lease do dono anterior expira
	if recuperado, err := engine.Recover(t.Context(), run.ID, "worker-2", time.Minute); err != nil ||
		recuperado.Status != wf.RunSucceeded {
		t.Fatalf("lease expirado deveria ser tomado: %v / %+v", err, recuperado)
	}
}

func TestRecoverVersaoIncompativelFalhaEstavel(t *testing.T) {
	engine, _, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	engine.WithRecovery(store)
	run := runInterrompido(t, store, 99) // versão não registrada
	recuperado, err := engine.Recover(t.Context(), run.ID, "worker-1", time.Minute)
	if !errors.Is(err, wf.ErrUnknownIdentity) || recuperado.Outcome != wf.OutcomeIncompatibleVersion {
		t.Fatalf("versão incompatível deveria falhar estável: %v / %+v", err, recuperado)
	}
}

func TestAbandonExpiredEncerraRunsParados(t *testing.T) {
	_, clock, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	run := runInterrompido(t, store, 1)
	corte := clock.Now().Add(time.Hour)
	encerrados, err := store.AbandonExpired(t.Context(), corte, corte)
	if err != nil || encerrados != 1 {
		t.Fatalf("limpeza deveria encerrar 1 run: %d (%v)", encerrados, err)
	}
	abandonado, _ := store.Get(t.Context(), run.ID)
	if abandonado.Status != wf.RunFailed || abandonado.Outcome != wf.OutcomeAbandoned {
		t.Fatalf("run parado deveria virar abandonado: %+v", abandonado)
	}
}
