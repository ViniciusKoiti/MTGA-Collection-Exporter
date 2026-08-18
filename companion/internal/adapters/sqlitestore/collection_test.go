package sqlitestore

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func snapshotFixture(t *testing.T, id collection.SnapshotID) collection.Snapshot {
	t.Helper()
	obs, err := collection.NewObservation(collection.SourceJSONImport, "export-1",
		time.Unix(1_699_999_000, 0).UTC(), nil, nil)
	if err != nil {
		t.Fatalf("observation: %v", err)
	}
	snap, err := collection.NewSnapshot(id, obs, obs.ObservedAt.Add(time.Minute),
		[]collection.Entry{
			{Identity: collection.CardIdentity{Printing: "p1", Arena: 101,
				Oracle: "o-1", Name: "Lightning Strike", Set: "DMU"}, Quantity: 4},
			{Unresolved: true, Raw: "??? unknown card", Quantity: 1},
		},
		[]collection.Diagnostic{{Code: "quantidade_zero", Detail: "line skipped",
			Severity: collection.SeverityInfo}})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

func TestSnapshotRoundTripPreservesEverything(t *testing.T) {
	store := abre(t)
	original := snapshotFixture(t, "snap-1")
	if err := store.Collections().Save(t.Context(), original); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := store.Collections().Get(t.Context(), "snap-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if loaded.Schema != original.Schema || loaded.Source != original.Source ||
		!loaded.ObservedAt.Equal(original.ObservedAt) ||
		len(loaded.Entries) != 2 || len(loaded.Diagnostics) != 1 {
		t.Fatalf("roundtrip diverged:\n got %+v\nwant %+v", loaded, original)
	}
	if loaded.Entries[0] != original.Entries[0] || loaded.Entries[1] != original.Entries[1] {
		t.Fatalf("entries diverged: %+v", loaded.Entries)
	}
	if loaded.TotalCartas() != 5 {
		t.Fatalf("totals diverged: %d", loaded.TotalCartas())
	}
}

func TestSnapshotsAreImmutableAndAtomic(t *testing.T) {
	store := abre(t)
	snap := snapshotFixture(t, "snap-1")
	if err := store.Collections().Save(t.Context(), snap); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.Collections().Save(t.Context(), snap); err == nil {
		t.Fatal("saving the same ID again must fail: snapshots are immutable")
	}
	loaded, err := store.Collections().Get(t.Context(), "snap-1")
	if err != nil || len(loaded.Entries) != 2 {
		t.Fatalf("failed re-save must not corrupt the stored snapshot: %+v (%v)", loaded, err)
	}
}

func TestLatestFollowsTheNewestSnapshotAcrossReopen(t *testing.T) {
	path := t.TempDir() + "/companion.db"
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := store.Collections().Save(t.Context(), snapshotFixture(t, "snap-1")); err != nil {
		t.Fatalf("save 1: %v", err)
	}
	if err := store.Collections().Save(t.Context(), snapshotFixture(t, "snap-2")); err != nil {
		t.Fatalf("save 2: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	latest, exists, err := reopened.Collections().Latest(t.Context())
	if err != nil || !exists || latest.ID != "snap-2" {
		t.Fatalf("latest should survive reopen as snap-2: %+v (%v)", latest, err)
	}
}
