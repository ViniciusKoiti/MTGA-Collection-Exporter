package diaglog

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func TestLocalLoggerScrubsSensitiveAttributes(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLocalLogger(&buffer)
	logger.Info("sync finished",
		"source_kind", "json_import",
		"export_path", `C:\Users\me\mtga_collection.json`,
		"auth", "Bearer abc123",
		"cards", 245)
	output := buffer.String()
	for _, leaked := range []string{"C:", "mtga_collection.json", "Bearer", "abc123"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("log leaked %q:\n%s", leaked, output)
		}
	}
	for _, kept := range []string{"json_import", "cards=245", "[redacted]"} {
		if !strings.Contains(output, kept) {
			t.Fatalf("log should keep %q:\n%s", kept, output)
		}
	}
}

func bundleDeps(t *testing.T) BundleDeps {
	t.Helper()
	snapshots := inmem.NewSnapshotStore()
	obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
		time.Unix(1_699_999_000, 0), nil, nil)
	snap, err := collection.NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute),
		[]collection.Entry{{Unresolved: true, Raw: "??? card", Quantity: 1}},
		[]collection.Diagnostic{{Code: "unknown_arena_id",
			Detail: `C:\Users\me\weird.json`, Severity: collection.SeverityWarning}})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := snapshots.Save(t.Context(), snap); err != nil {
		t.Fatalf("seed: %v", err)
	}
	audit := &inmem.AuditLog{}
	_ = audit.Append(t.Context(), approvals.AuditRecord{Correlation: "tok-1",
		Tool: "copy-export", ArgsHash: "abcd", Outcome: approvals.AuditExecuted,
		At: time.Unix(1_700_000_000, 0)})
	return BundleDeps{Snapshots: snapshots, Audit: audit}
}

func TestBundleHonorsOptInSections(t *testing.T) {
	deps := bundleDeps(t)
	minimal, err := ExportBundle(t.Context(), BundleOptions{}, deps)
	if err != nil || strings.Count(minimal, "\n") != 1 {
		t.Fatalf("empty opt-in must yield header only: %q (%v)", minimal, err)
	}
	full, err := ExportBundle(t.Context(), BundleOptions{
		IncludeSyncStatus: true, IncludeDiagnostics: true, IncludeActivity: true,
	}, deps)
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	for _, expected := range []string{"sync|ready|snapshot=snap-1",
		"diagnostic|warning|unknown_arena_id", "activity|copy-export|executed|tok-1"} {
		if !strings.Contains(full, expected) {
			t.Fatalf("bundle missing %q:\n%s", expected, full)
		}
	}
	if strings.Contains(full, "C:") || strings.Contains(full, "weird.json") {
		t.Fatalf("bundle leaked a path:\n%s", full)
	}
	syncOnly, _ := ExportBundle(t.Context(), BundleOptions{IncludeSyncStatus: true}, deps)
	if strings.Contains(syncOnly, "activity|") || strings.Contains(syncOnly, "diagnostic|") {
		t.Fatalf("sections without opt-in must be absent:\n%s", syncOnly)
	}
}
