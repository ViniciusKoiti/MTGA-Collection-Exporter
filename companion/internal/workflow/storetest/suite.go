// Package storetest é a suíte de contrato reutilizável de RunStore
// (tarefa 3.5 do OpenSpec add-graph-workflow-harness): todo adapter —
// in-memory, SQLite e, futuramente, PostgreSQL — deve passar exatamente
// nos mesmos casos.
package storetest

import (
	"encoding/json"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Fabrica cria um RunStore novo e vazio para cada subteste.
type Fabrica func(t *testing.T) wf.RunStore

// Executa roda a suíte de contrato completa contra a fábrica dada.
func Executa(t *testing.T, novo Fabrica) {
	t.Run("CreateEGetPreservamCheckpoint", func(t *testing.T) { createEGet(t, novo(t)) })
	t.Run("CreateDuplicadoFalha", func(t *testing.T) { createDuplicado(t, novo(t)) })
	t.Run("UpdateExigeVersaoSequencial", func(t *testing.T) { versaoOtimista(t, novo(t)) })
	t.Run("RunInexistenteFalha", func(t *testing.T) { inexistente(t, novo(t)) })
	t.Run("StepsPreservamOrdem", func(t *testing.T) { stepsOrdenados(t, novo(t)) })
}

func runExemplo() wf.Run {
	return wf.Run{
		ID:        "run-000001",
		Graph:     wf.Identity{Kind: "collection-sync", Version: 1},
		Status:    wf.RunActive,
		Current:   "coleta",
		State:     map[string]any{"cartas": "42"},
		Version:   1,
		StartedAt: time.Unix(1_700_000_000, 0).UTC(),
		UpdatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}
}

// mesmoEstado compara estados pela forma canônica JSON, já que adapters
// persistentes normalizam tipos concretos no roundtrip.
func mesmoEstado(t *testing.T, a, b wf.State) bool {
	t.Helper()
	ja, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jb, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(ja) == string(jb)
}

func createEGet(t *testing.T, store wf.RunStore) {
	run := runExemplo()
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
	lido, err := store.Get(t.Context(), run.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if lido.ID != run.ID || lido.Graph != run.Graph || lido.Status != run.Status ||
		lido.Current != run.Current || lido.Version != run.Version ||
		!lido.StartedAt.Equal(run.StartedAt) || !mesmoEstado(t, lido.State, run.State) {
		t.Fatalf("roundtrip divergente:\n got %+v\nwant %+v", lido, run)
	}
}

func createDuplicado(t *testing.T, store wf.RunStore) {
	if err := store.Create(t.Context(), runExemplo()); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := store.Create(t.Context(), runExemplo()); err == nil {
		t.Fatal("segundo create do mesmo run deveria falhar")
	}
}

func inexistente(t *testing.T, store wf.RunStore) {
	if _, err := store.Get(t.Context(), "run-999999"); err == nil {
		t.Fatal("get de run inexistente deveria falhar")
	}
	fantasma := runExemplo()
	fantasma.ID = "run-999999"
	fantasma.Version = 2
	if err := store.Update(t.Context(), fantasma); err == nil {
		t.Fatal("update de run inexistente deveria falhar")
	}
}
