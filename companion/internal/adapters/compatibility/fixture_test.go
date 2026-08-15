package compatibility

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func snapshotFixture(t *testing.T) collection.Snapshot {
	t.Helper()
	observedAt := time.Unix(1_700_000_000, 0).UTC()
	observation, err := collection.NewObservation(collection.SourceJSONImport,
		"fixture", observedAt, nil, nil)
	if err != nil {
		t.Fatalf("observation: %v", err)
	}
	entries := []collection.Entry{
		{Identity: collection.CardIdentity{Printing: "p2", Arena: 20,
			Name: "Zeta Card", Set: "TST"}, Quantity: 1},
		{Unresolved: true, Raw: "private unknown", Quantity: 3},
		{Identity: collection.CardIdentity{Printing: "p1", Arena: 10,
			Name: "Alpha Card", Set: "ONE"}, Quantity: 2},
		{Identity: collection.CardIdentity{Printing: "p1b", Arena: 11,
			Name: "Alpha Card", Set: "ONE"}, Quantity: 2},
	}
	snapshot, err := collection.NewSnapshot("snap-1", observation,
		observedAt.Add(time.Minute), entries, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snapshot
}
