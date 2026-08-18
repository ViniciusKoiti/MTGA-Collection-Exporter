package sqlitestore

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/storetest"
)

func abre(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "runs.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// O adapter SQLite passa na mesma suíte de contrato que o in-memory (3.5).
func TestRunStoreSQLiteCumpreContrato(t *testing.T) {
	storetest.Executa(t, func(t *testing.T) wf.RunStore { return abre(t) })
}

// Migrações são idempotentes: reabrir o mesmo banco não reaplica nada.
func TestMigracoesSaoIdempotentes(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "runs.db")
	primeiro, err := Open(caminho)
	if err != nil {
		t.Fatalf("primeira abertura: %v", err)
	}
	run := wf.Run{ID: "run-000001", Graph: wf.Identity{Kind: "collection-sync", Version: 1},
		Status: wf.RunActive, Current: "coleta", Version: 1,
		StartedAt: time.Unix(1_700_000_000, 0).UTC(), UpdatedAt: time.Unix(1_700_000_000, 0).UTC()}
	if err := primeiro.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := primeiro.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	segundo, err := Open(caminho)
	if err != nil {
		t.Fatalf("reabertura deveria reaproveitar o schema: %v", err)
	}
	defer func() { _ = segundo.Close() }()
	if _, err := segundo.Get(t.Context(), run.ID); err != nil {
		t.Fatalf("dados deveriam sobreviver à reabertura: %v", err)
	}
}

// Checkpoint e outbox são atômicos: conflito de versão desfaz os efeitos.
func TestCheckpointComEfeitosEhAtomico(t *testing.T) {
	store := abre(t)
	run := wf.Run{ID: "run-000001", Graph: wf.Identity{Kind: "approved-export", Version: 1},
		Status: wf.RunActive, Current: "exporta", Version: 1,
		StartedAt: time.Unix(1_700_000_000, 0).UTC(), UpdatedAt: time.Unix(1_700_000_000, 0).UTC()}
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
	efeito := wf.EffectRecord{ID: "efeito-1", Run: run.ID,
		Preview:    wf.EffectPreview{Effect: "export-write", Target: "decks/a.txt", PayloadHash: "h"},
		EnqueuedAt: run.StartedAt}
	conflito := run
	conflito.Version = 9 // versão errada: transação inteira deve reverter
	err := store.CheckpointComEfeitos(t.Context(), conflito, []wf.EffectRecord{efeito})
	if !errors.Is(err, wf.ErrVersionConflict) {
		t.Fatalf("esperava conflito de versão, veio: %v", err)
	}
	pendentes, _ := store.EfeitosPendentes(t.Context())
	if len(pendentes) != 0 {
		t.Fatalf("rollback deveria descartar os efeitos: %+v", pendentes)
	}
	valido := run
	valido.Version = 2
	if err := store.CheckpointComEfeitos(t.Context(), valido,
		[]wf.EffectRecord{efeito, efeito}); err != nil {
		t.Fatalf("checkpoint válido: %v", err)
	}
	pendentes, _ = store.EfeitosPendentes(t.Context())
	if len(pendentes) != 1 || pendentes[0].ID != "efeito-1" {
		t.Fatalf("efeito duplicado deveria deduplicar: %+v", pendentes)
	}
	if err := store.AckEfeito(t.Context(), "efeito-1"); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if pendentes, _ = store.EfeitosPendentes(t.Context()); len(pendentes) != 0 {
		t.Fatalf("ack deveria limpar pendências: %+v", pendentes)
	}
}
