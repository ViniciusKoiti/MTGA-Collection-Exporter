// Parity trial (OpenSpec introduce-agentic-go-companion, task 7.5): feed
// REAL Python exporter snapshots through the Go source and the single
// product normalization and prove card-for-card quantity parity. Gated by
// COMPANION_PARITY_DIR because the snapshots are user data, not fixtures.
package legacyjson_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/legacyjson"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/normalize"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Unix(1_700_000_000, 0).UTC() }

// noCatalog resolves nothing on purpose: the trial pins SOURCE parity,
// not catalog coverage — every entry stays unresolved and must still
// carry the exact Python quantity under the exact raw identity.
type noCatalog struct{}

func (noCatalog) ResolvePorArena(context.Context, collection.ArenaID) (
	collection.CardIdentity, bool, error) {
	return collection.CardIdentity{}, false, nil
}

func (noCatalog) ResolvePorNome(context.Context, string) (
	collection.CardIdentity, bool, error) {
	return collection.CardIdentity{}, false, nil
}

// pythonRow is an INDEPENDENT minimal read of the export; the trial must
// not reuse the adapter's decoder to count the Python side.
type pythonRow struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Set   string `json:"set"`
}

func TestPythonExportsAndGoNormalizationAgreeCardForCard(t *testing.T) {
	dir := os.Getenv("COMPANION_PARITY_DIR")
	if dir == "" {
		t.Skip("set COMPANION_PARITY_DIR to a folder of real mtga_collection.json snapshots")
	}
	snapshots, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(snapshots) < 2 {
		t.Fatalf("the trial needs snapshots from at least two moments: %v (%v)", snapshots, err)
	}
	for _, path := range snapshots {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read snapshot: %v", err)
			}
			var python []pythonRow
			if err := json.Unmarshal(raw, &python); err != nil {
				t.Fatalf("independent python read: %v", err)
			}
			want, wantTotal, zeroRows := map[string]int{}, 0, 0
			for _, row := range python {
				if row.Count == 0 {
					zeroRows++
					continue
				}
				want[fmt.Sprintf("%s (%s)", row.Name, row.Set)] += row.Count
				wantTotal += row.Count
			}
			source := &legacyjson.Source{Path: path, Clock: fixedClock{}}
			obs, err := source.Observe(context.Background())
			if err != nil {
				t.Fatalf("go source rejected a real export: %v", err)
			}
			result, err := normalize.Observation(context.Background(), noCatalog{}, obs)
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			got, gotTotal := map[string]int{}, 0
			for _, entry := range result.Entries {
				got[entry.Raw] += entry.Quantity
				gotTotal += entry.Quantity
			}
			if gotTotal != wantTotal || len(result.Entries) != len(python)-zeroRows {
				t.Fatalf("totals drifted: go %d cards/%d entries, python %d cards/%d rows",
					gotTotal, len(result.Entries), wantTotal, len(python)-zeroRows)
			}
			for identity, quantity := range want {
				if got[identity] != quantity {
					t.Errorf("%s: go has %d, python has %d", identity, got[identity], quantity)
				}
			}
			t.Logf("parity ok: %d identities, %d cards, %d zero rows", len(want), wantTotal, zeroRows)
		})
	}
}
