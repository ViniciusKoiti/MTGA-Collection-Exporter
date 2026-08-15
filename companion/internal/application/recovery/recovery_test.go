package recovery

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/compatibility"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/legacyjson"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/sqlitestore"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// TestDatabaseLossRecoversFromCompatibilityExport is the end-to-end drill
// of task 7.2: snapshot in SQLite -> compatibility export -> database is
// lost -> a FRESH database is rebuilt from the export through the frozen
// legacy contract, preserving resolved totals.
func TestDatabaseLossRecoversFromCompatibilityExport(t *testing.T) {
	ctx := t.Context()
	clock := inmem.NewClock(time.Unix(1_700_000_000, 0))
	exportDir := t.TempDir()

	// Original database with one committed snapshot.
	original, err := sqlitestore.Open(filepath.Join(t.TempDir(), "companion.db"))
	if err != nil {
		t.Fatalf("open original: %v", err)
	}
	obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
		clock.Now().Add(-time.Hour), nil, nil)
	identity := collection.CardIdentity{Printing: "p1", Arena: 82183,
		Name: "Lightning Strike", Set: "DMU"}
	snap, err := collection.NewSnapshot("snap-1", obs, clock.Now(),
		[]collection.Entry{
			{Identity: identity, Quantity: 4},
			{Unresolved: true, Raw: "??? mystery", Quantity: 1},
		}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := original.Collections().Save(ctx, snap); err != nil {
		t.Fatalf("seed: %v", err)
	}
	paths, err := compatibility.Writer{Exporter: compatibility.Exporter{}}.
		WriteAll(ctx, exportDir, snap)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := original.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	// Database lost: nothing restores it; only the export survives.

	fresh, err := sqlitestore.Open(filepath.Join(t.TempDir(), "rebuilt.db"))
	if err != nil {
		t.Fatalf("open fresh: %v", err)
	}
	defer func() { _ = fresh.Close() }()
	source := &legacyjson.Source{
		Path:  paths[ports.ExportJSON],
		Clock: clock,
	}
	recovered, err := FromExport(ctx, source,
		&inmem.Catalog{PorArena: map[collection.ArenaID]collection.CardIdentity{
			82183: identity,
		}}, fresh.Collections(), clock)
	if err != nil {
		t.Fatalf("recovery: %v", err)
	}
	if recovered.TotalCartas() != 4 { // unresolved had no legacy representation
		t.Fatalf("recovered totals diverged: %d", recovered.TotalCartas())
	}
	if !strings.HasPrefix(string(recovered.ID), "recovered-") ||
		strings.Contains(string(recovered.ID), "\\") {
		t.Fatalf("recovered ID must be marked and path-free: %s", recovered.ID)
	}
	latest, exists, err := fresh.Collections().Latest(ctx)
	if err != nil || !exists || latest.ID != recovered.ID {
		t.Fatalf("recovered snapshot must be the new latest: %+v (%v)", latest, err)
	}
	marked := false
	for _, diag := range latest.Diagnostics {
		if diag.Code == "recovered_from_export" {
			marked = true
		}
	}
	if !marked {
		t.Fatal("recovery provenance diagnostic missing")
	}
}
