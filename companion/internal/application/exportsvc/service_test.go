package exportsvc

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

type fixture struct {
	service   *Service
	snapshots *inmem.SnapshotStore
	clipboard *inmem.Clipboard
	audit     *inmem.AuditLog
	clock     *inmem.Clock
}

func setup(t *testing.T) fixture {
	t.Helper()
	clock := inmem.NewClock(time.Unix(1_700_000_000, 0))
	snapshots := inmem.NewSnapshotStore()
	seed(t, snapshots, "snap-1", 4)
	clipboard := &inmem.Clipboard{}
	audit := &inmem.AuditLog{}
	service := New(snapshots, inmem.Exporter{}, clipboard,
		inmem.NewApprovalService(clock), audit, clock)
	return fixture{service, snapshots, clipboard, audit, clock}
}

func seed(t *testing.T, store *inmem.SnapshotStore, id collection.SnapshotID, qty int) {
	t.Helper()
	obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
		time.Unix(1_699_999_000, 0), nil, nil)
	snap, err := collection.NewSnapshot(id, obs, obs.ObservedAt.Add(time.Minute),
		[]collection.Entry{{Identity: collection.CardIdentity{
			Printing: "p1", Arena: 101, Name: "Card A", Set: "TST"}, Quantity: qty}}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := store.Save(t.Context(), snap); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func TestDirectUserCopyExecutesImmediately(t *testing.T) {
	f := setup(t)
	bytes, err := f.service.CopyToClipboard(t.Context(), ports.ExportText)
	if err != nil || bytes == 0 || len(f.clipboard.Textos) != 1 {
		t.Fatalf("direct copy failed: %d bytes (%v), clipboard %v", bytes, err, f.clipboard.Textos)
	}
	trail, _ := f.audit.Recent(t.Context(), 10)
	if len(trail) != 1 || trail[0].Outcome != "executed" || trail[0].Correlation != "user" {
		t.Fatalf("audit trail unexpected: %+v", trail)
	}
}

func TestAssistantProposeThenApprovedExecution(t *testing.T) {
	f := setup(t)
	proposal, err := f.service.ProposeCopy(t.Context(), ports.ExportJSON)
	if err != nil || proposal.Snapshot != "snap-1" || proposal.Bytes == 0 {
		t.Fatalf("proposal failed: %+v (%v)", proposal, err)
	}
	if len(f.clipboard.Textos) != 0 {
		t.Fatal("proposal alone must not touch the clipboard")
	}
	if _, err := f.service.ExecuteApproved(t.Context(), proposal.TokenID, ports.ExportJSON); err != nil {
		t.Fatalf("approved execution failed: %v", err)
	}
	if len(f.clipboard.Textos) != 1 {
		t.Fatal("approved execution should copy exactly once")
	}
	if _, err := f.service.ExecuteApproved(t.Context(), proposal.TokenID, ports.ExportJSON); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("token reuse must be denied: %v", err)
	}
	trail, _ := f.audit.Recent(t.Context(), 10)
	outcomes := []string{}
	for _, rec := range trail {
		outcomes = append(outcomes, string(rec.Outcome))
	}
	if len(outcomes) != 3 || outcomes[0] != "requested" || outcomes[1] != "executed" ||
		outcomes[2] != "denied" {
		t.Fatalf("audit sequence unexpected: %v", outcomes)
	}
}
