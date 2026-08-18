package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

func manifestFor(key string) catalog.Manifest {
	return catalog.Manifest{KeyID: "key-1", Signature: "sig",
		Object: catalog.ObjectRef{Key: key, SHA256: "sha-" + key, Size: 10}}
}

// TestActivationIsAtomicWithPreviousFallback: the current pointer
// moves atomically, failures leave it untouched, and rollback restores
// the previous snapshot.
func TestActivationIsAtomicWithPreviousFallback(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO sources (id, kind, approved)
		VALUES ('scryfall', 'cards', TRUE)`); err != nil {
		t.Fatalf("seed source: %v", err)
	}
	act := CatalogActivator{Pool: pool, SourceID: "scryfall",
		SchemaName: "cards-v1"}

	if err := act.Activate(ctx, manifestFor("snap-a")); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if id, err := act.CurrentArtifactID(ctx); err != nil || id != "snap-a" {
		t.Fatalf("snap-a must be current: %q %v", id, err)
	}
	if err := act.Activate(ctx, manifestFor("snap-b")); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	if id, _ := act.CurrentArtifactID(ctx); id != "snap-b" {
		t.Fatalf("activation must swap to snap-b, got %q", id)
	}
	if err := act.Activate(ctx, manifestFor("snap-b")); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate activation must fail cleanly: %v", err)
	}
	if id, _ := act.CurrentArtifactID(ctx); id != "snap-b" {
		t.Fatalf("failed activation must leave the current untouched: %q", id)
	}
	if err := act.RollbackToPrevious(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if id, _ := act.CurrentArtifactID(ctx); id != "snap-a" {
		t.Fatalf("rollback must restore the previous snapshot: %q", id)
	}
}
