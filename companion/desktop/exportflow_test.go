//go:build windows

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func seedSnapshot(t *testing.T) {
	t.Helper()
	// Not t.TempDir(): the package-level store singleton keeps the
	// database open past the test, which would fail the auto-cleanup.
	dir, err := os.MkdirTemp("", "companion-export-flow-")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Setenv("COMPANION_DATA_DIR", dir)
	snapshots, err := collectionsStore()
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	obs, err := collection.NewObservation(collection.SourceJSONImport,
		"seed", time.Unix(1_700_000_000, 0).UTC(), nil, nil)
	if err != nil {
		t.Fatalf("observation: %v", err)
	}
	snap, err := collection.NewSnapshot("seed-1", obs,
		obs.ObservedAt.Add(time.Minute), []collection.Entry{{
			Identity: collection.CardIdentity{Printing: "p1", Arena: 101,
				Oracle: "o1", Name: "Synthetic Bolt", Set: "TST"},
			Quantity: 4}}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := snapshots.Save(context.Background(), snap); err != nil {
		t.Fatalf("save: %v", err)
	}
}

// TestApprovedExportPausesAndDenialCarriesItsStableCode pins the
// desktop approval flow end to end at the binding level.
func TestApprovedExportPausesAndDenialCarriesItsStableCode(t *testing.T) {
	seedSnapshot(t)
	app := NewApp()
	preview := app.StartApprovedExport("json")
	if preview.State.Status != viewstate.StatusStale || preview.RunID == "" ||
		preview.DecisionKey == "" || preview.Bytes == 0 {
		t.Fatalf("the run must pause on the preview: %+v", preview)
	}
	denied := app.DecideExport(preview.RunID, preview.DecisionKey, false)
	if denied.State.Status != viewstate.StatusError ||
		denied.State.Code != apperr.CodeApprovalDenied {
		t.Fatalf("denial must carry its stable code: %+v", denied.State)
	}
	if denied.Written != "" {
		t.Fatalf("a denied export must write nothing: %q", denied.Written)
	}
	approvedPreview := app.StartApprovedExport("json")
	approved := app.DecideExport(approvedPreview.RunID,
		approvedPreview.DecisionKey, true)
	if approved.State.Status != viewstate.StatusSuccess ||
		approved.Written == "" {
		t.Fatalf("an approved export must write: %+v", approved)
	}
}
