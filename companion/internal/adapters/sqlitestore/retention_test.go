package sqlitestore

import (
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func runComEventos(t *testing.T, store *Store, id wf.RunID, status wf.RunStatus, em time.Time, eventos int) {
	t.Helper()
	run := wf.Run{ID: id, Graph: wf.Identity{Kind: "collection-sync", Version: 1},
		Status: status, Current: "coleta", Version: 1, StartedAt: em, UpdatedAt: em}
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create %s: %v", id, err)
	}
	for i := range eventos {
		ev := wf.Event{Schema: wf.EventSchema, Run: id, Step: i + 1,
			Graph:       wf.Identity{Kind: "collection-sync", Version: 1},
			Correlation: string(id), At: em, Outcome: "ok"}
		if err := store.Emit(t.Context(), ev); err != nil {
			t.Fatalf("emit %s/%d: %v", id, i, err)
		}
	}
}

func contaEventos(t *testing.T, store *Store, id wf.RunID) int {
	t.Helper()
	eventos, _, err := store.Timeline(t.Context(), id, 0, 1000)
	if err != nil {
		t.Fatalf("timeline %s: %v", id, err)
	}
	return len(eventos)
}

func TestRetencaoPorIdadePreservaCheckpointsAtivos(t *testing.T) {
	store := abre(t)
	antigo := time.Unix(1_700_000_000, 0).UTC()
	recente := antigo.Add(48 * time.Hour)
	runComEventos(t, store, "run-velho-ok", wf.RunSucceeded, antigo, 2)
	runComEventos(t, store, "run-ativo-velho", wf.RunActive, antigo, 2)
	runComEventos(t, store, "run-novo-ok", wf.RunSucceeded, recente, 2)

	relatorio, err := store.ApplyRetention(t.Context(), antigo.Add(24*time.Hour), 1000)
	if err != nil {
		t.Fatalf("retenção: %v", err)
	}
	if relatorio.RunsRemovidos != 1 {
		t.Fatalf("apenas o terminal antigo deveria sair: %+v", relatorio)
	}
	if _, err := store.Get(t.Context(), "run-velho-ok"); err == nil {
		t.Fatal("run terminal antigo deveria ter sido removido")
	}
	if _, err := store.Get(t.Context(), "run-ativo-velho"); err != nil {
		t.Fatalf("checkpoint ativo antigo é intocável: %v", err)
	}
	if contaEventos(t, store, "run-ativo-velho") != 2 || contaEventos(t, store, "run-novo-ok") != 2 {
		t.Fatal("eventos de runs ativos e recentes devem sobreviver à idade")
	}
}

func TestRetencaoPorTamanhoRemoveApenasJournalTerminal(t *testing.T) {
	store := abre(t)
	base := time.Unix(1_700_000_000, 0).UTC()
	runComEventos(t, store, "run-terminal", wf.RunSucceeded, base, 4)
	runComEventos(t, store, "run-ativo", wf.RunActive, base, 2)

	relatorio, err := store.ApplyRetention(t.Context(), base.Add(-time.Hour), 3)
	if err != nil {
		t.Fatalf("retenção: %v", err)
	}
	if relatorio.EventosRemovidos != 3 { // 6 no total, teto 3, só terminais saem
		t.Fatalf("esperava 3 eventos removidos: %+v", relatorio)
	}
	if contaEventos(t, store, "run-ativo") != 2 {
		t.Fatal("journal de run ativo não pode ser podado por tamanho")
	}
	if contaEventos(t, store, "run-terminal") != 1 {
		t.Fatal("os eventos terminais mais antigos deveriam sair primeiro")
	}
}
