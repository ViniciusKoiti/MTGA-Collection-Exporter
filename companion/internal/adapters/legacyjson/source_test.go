package legacyjson

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func source(name string) *Source {
	return &Source{
		Path:  filepath.Join("testdata", name),
		Clock: inmem.NewClock(time.Unix(1_700_000_000, 0)),
	}
}

// TestValidExportContract freezes the happy-path contract: duplicate
// printings stay as separate rows and the unknown card survives as a raw
// record with a warning diagnostic.
func TestValidExportContract(t *testing.T) {
	obs, err := source("valid_export.json").Observe(t.Context())
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if obs.Source != collection.SourceJSONImport || len(obs.Quantities) != 3 {
		t.Fatalf("observation unexpected: %+v", obs)
	}
	first, second, unknown := obs.Quantities[0], obs.Quantities[1], obs.Quantities[2]
	if first.Arena != 82183 || first.Quantity != 4 ||
		first.RawIdentity != "Lightning Strike (DMU)" {
		t.Fatalf("first printing unexpected: %+v", first)
	}
	if second.Arena != 71234 || second.RawIdentity != "Lightning Strike (M21)" {
		t.Fatalf("duplicate printing must stay separate: %+v", second)
	}
	if unknown.Arena != 0 || unknown.Quantity != 1 {
		t.Fatalf("unknown card must stay observable: %+v", unknown)
	}
	if len(obs.Diagnostics) != 1 || obs.Diagnostics[0].Code != "unknown_arena_id" ||
		obs.Diagnostics[0].Severity != collection.SeverityWarning {
		t.Fatalf("unknown card needs a warning diagnostic: %+v", obs.Diagnostics)
	}
}

func TestEmptyExportIsAValidObservation(t *testing.T) {
	obs, err := source("empty_export.json").Observe(t.Context())
	if err != nil || len(obs.Quantities) != 0 || len(obs.Diagnostics) != 0 {
		t.Fatalf("empty export should observe cleanly: %+v (%v)", obs, err)
	}
}

func TestContractRejectsDriftAndDamage(t *testing.T) {
	cases := map[string]string{
		"malformed export":      "malformed_export.json",
		"unknown field (drift)": "unknown_field_export.json",
		"missing file":          "does_not_exist.json",
	}
	for name, fixture := range cases {
		if _, err := source(fixture).Observe(t.Context()); err == nil {
			t.Errorf("%s: must fail loudly", name)
		}
	}
}
