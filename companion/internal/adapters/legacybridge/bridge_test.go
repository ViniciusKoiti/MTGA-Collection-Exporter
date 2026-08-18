package legacybridge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// TestHelperProcess is the fake legacy scanner: re-executing the test
// binary keeps the suite portable (no cmd/bash dependency).
func TestHelperProcess(t *testing.T) {
	if os.Getenv("LEGACY_HELPER") != "1" {
		return
	}
	switch os.Getenv("HELPER_MODE") {
	case "ok":
		payload := `[{"count":4,"name":"Lightning Strike","set":"DMU",
			"collector_number":"137","rarity":"","colors":[],"type_line":"",
			"mana_cost":"","cmc":null,"image":"","grp_id":82183}]`
		_ = os.WriteFile(filepath.Join(os.Getenv("HELPER_OUT"), "mtga_collection.json"),
			[]byte(payload), 0o644)
		os.Exit(0)
	case "hang":
		time.Sleep(30 * time.Second)
		os.Exit(0)
	default:
		os.Exit(3)
	}
}

func bridge(t *testing.T, mode string, timeout time.Duration) (*Bridge, string) {
	t.Helper()
	dir := t.TempDir()
	return &Bridge{
		Command:   []string{os.Args[0], "-test.run=TestHelperProcess"},
		Env:       []string{"LEGACY_HELPER=1", "HELPER_MODE=" + mode, "HELPER_OUT=" + dir},
		OutputDir: dir,
		Timeout:   timeout,
		Clock:     inmem.NewClock(time.Unix(1_700_000_000, 0)),
	}, dir
}

func TestBridgeRunsScannerAndRelabelsProvenance(t *testing.T) {
	source, _ := bridge(t, "ok", 20*time.Second)
	obs, err := source.Observe(t.Context())
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if obs.Source != collection.SourceLegacyBridge {
		t.Fatalf("provenance must be the bridge, got %s", obs.Source)
	}
	if len(obs.Quantities) != 1 || obs.Quantities[0].Arena != 82183 ||
		obs.Quantities[0].Quantity != 4 {
		t.Fatalf("imported export unexpected: %+v", obs.Quantities)
	}
}

func TestBridgeFailuresAreLoud(t *testing.T) {
	failing, _ := bridge(t, "fail", 20*time.Second)
	if _, err := failing.Observe(t.Context()); err == nil {
		t.Fatal("non-zero scanner exit must fail")
	}
	silent, _ := bridge(t, "silent", 20*time.Second) // exits 3, writes nothing
	if _, err := silent.Observe(t.Context()); err == nil {
		t.Fatal("missing export must fail")
	}
	unconfigured := &Bridge{OutputDir: t.TempDir()}
	if _, err := unconfigured.Observe(t.Context()); err == nil {
		t.Fatal("no command configured must fail")
	}
}

func TestBridgeDeadlineKillsHangingScanner(t *testing.T) {
	hanging, _ := bridge(t, "hang", 2*time.Second)
	start := time.Now()
	_, err := hanging.Observe(t.Context())
	if err == nil || errors.Is(err, context.Canceled) {
		t.Fatalf("hang must be killed by the deadline: %v", err)
	}
	if time.Since(start) > 15*time.Second {
		t.Fatal("deadline was not enforced")
	}
}
