package diaglog

import (
	"context"
	"fmt"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// BundleOptions are the EXPLICIT opt-in fields of the diagnostic bundle:
// nothing is included unless the user turned the section on.
type BundleOptions struct {
	IncludeSyncStatus  bool
	IncludeDiagnostics bool
	IncludeActivity    bool
	ActivityLimit      int
}

// BundleDeps are the read-only sources the exporter may consult.
type BundleDeps struct {
	Snapshots ports.SnapshotStore
	Audit     ports.AuditLog
}

// ExportBundle renders the redacted diagnostic bundle. Every value is
// scrubbed and the format is line-oriented and stable for support use.
func ExportBundle(ctx context.Context, opts BundleOptions, deps BundleDeps) (string, error) {
	var b strings.Builder
	b.WriteString("bundle|companion-diagnostics/v1\n")
	if opts.IncludeSyncStatus || opts.IncludeDiagnostics {
		snap, exists, err := deps.Snapshots.Latest(ctx)
		if err != nil {
			return "", err
		}
		if opts.IncludeSyncStatus {
			if !exists {
				b.WriteString("sync|not_configured\n")
			} else {
				fmt.Fprintf(&b, "sync|ready|snapshot=%s|source=%s|cards=%d|entries=%d\n",
					scrub(string(snap.ID)), scrub(string(snap.Source)),
					snap.TotalCartas(), len(snap.Entries))
			}
		}
		if opts.IncludeDiagnostics && exists {
			for _, diag := range snap.Diagnostics {
				fmt.Fprintf(&b, "diagnostic|%s|%s|%s\n",
					diag.Severity, scrub(diag.Code), scrub(diag.Detail))
			}
		}
	}
	if opts.IncludeActivity {
		limit := opts.ActivityLimit
		if limit <= 0 {
			limit = 20
		}
		records, err := deps.Audit.Recent(ctx, limit)
		if err != nil {
			return "", err
		}
		for _, rec := range records {
			fmt.Fprintf(&b, "activity|%s|%s|%s\n",
				scrub(rec.Tool), outcomeCode(rec.Outcome), scrub(rec.Correlation))
		}
	}
	return b.String(), nil
}

func outcomeCode(outcome approvals.AuditOutcome) string {
	return scrub(string(outcome))
}
