package parity

// Goldens in testdata/ were captured from mtga/mcp_server.py
// (collection_stats and check_deck) over the shared fixture
// companion/internal/adapters/legacyjson/testdata/valid_export.json.
// To regenerate after an intentional Python change, run the snippet in
// docs/activity-inventory.md's parity note (monkeypatch _load with the
// fixture and dump both results).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/legacyjson"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/normalize"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func sharedSnapshot(t *testing.T) collection.Snapshot {
	t.Helper()
	source := &legacyjson.Source{
		Path:  filepath.Join("..", "adapters", "legacyjson", "testdata", "valid_export.json"),
		Clock: inmem.NewClock(time.Unix(1_700_000_000, 0)),
	}
	obs, err := source.Observe(t.Context())
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	catalog := &inmem.Catalog{PorArena: map[collection.ArenaID]collection.CardIdentity{
		82183: {Printing: "p1", Arena: 82183, Name: "Lightning Strike", Set: "DMU"},
		71234: {Printing: "p2", Arena: 71234, Name: "Lightning Strike", Set: "M21"},
	}}
	result, err := normalize.Observation(t.Context(), catalog, obs)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	snap, err := collection.NewSnapshot("parity", obs, obs.ObservedAt.Add(time.Minute),
		result.Entries, result.Diagnostics)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

func golden(t *testing.T, name string, out any) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("golden %s: %v", name, err)
	}
	if err := json.Unmarshal(payload, out); err != nil {
		t.Fatalf("parse golden %s: %v", name, err)
	}
}

func TestCollectionTotalsMatchThePythonMCP(t *testing.T) {
	var stats struct {
		TotalCopies int `json:"total_copies"`
		UniqueCards int `json:"unique_cards"`
	}
	golden(t, "python_stats.golden.json", &stats)
	snap := sharedSnapshot(t)
	if snap.TotalCartas() != stats.TotalCopies {
		t.Fatalf("total copies diverged: Go %d vs Python %d",
			snap.TotalCartas(), stats.TotalCopies)
	}
	if len(snap.Entries) != stats.UniqueCards {
		t.Fatalf("row count diverged: Go %d vs Python %d",
			len(snap.Entries), stats.UniqueCards)
	}
}
