package sqlitestore

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func snapshotAt(t *testing.T, id collection.SnapshotID,
	importedAt time.Time) collection.Snapshot {
	t.Helper()
	obs, err := collection.NewObservation(collection.SourceJSONImport,
		"export-"+string(id), importedAt.Add(-time.Minute).UTC(), nil, nil)
	if err != nil {
		t.Fatalf("observation: %v", err)
	}
	snap, err := collection.NewSnapshot(id, obs, importedAt.UTC(),
		[]collection.Entry{{Identity: collection.CardIdentity{Printing: "p1",
			Arena: 101, Oracle: "o-1", Name: "Synthetic Bolt", Set: "TST"},
			Quantity: 4}}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

// TestPreviousReturnsTheSecondNewestSnapshot: zero or one snapshot
// means no previous (a normal state), and with two the older one
// backs the Home delta.
func TestPreviousReturnsTheSecondNewestSnapshot(t *testing.T) {
	store := abre(t)
	ctx := t.Context()
	if _, ok, err := store.Collections().Previous(ctx); err != nil || ok {
		t.Fatalf("no snapshots means no previous: %v %v", ok, err)
	}
	base := time.Unix(1_700_000_000, 0)
	if err := store.Collections().Save(ctx,
		snapshotAt(t, "snap-a", base)); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if _, ok, err := store.Collections().Previous(ctx); err != nil || ok {
		t.Fatalf("one snapshot still has no previous: %v %v", ok, err)
	}
	if err := store.Collections().Save(ctx,
		snapshotAt(t, "snap-b", base.Add(time.Hour))); err != nil {
		t.Fatalf("save b: %v", err)
	}
	previous, ok, err := store.Collections().Previous(ctx)
	if err != nil || !ok || previous.ID != "snap-a" {
		t.Fatalf("the older snapshot must back the delta: %+v %v %v",
			previous, ok, err)
	}
}
