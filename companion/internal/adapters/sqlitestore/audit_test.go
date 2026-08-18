package sqlitestore

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
)

func auditRecord(correlation string, outcome approvals.AuditOutcome, at time.Time) approvals.AuditRecord {
	return approvals.AuditRecord{Correlation: correlation, Tool: "copy-export",
		ArgsHash: "abcd1234", Outcome: outcome, At: at}
}

func TestAuditTrailAndActivityView(t *testing.T) {
	store := abre(t)
	base := time.Unix(1_700_000_000, 0).UTC()
	sequence := []approvals.AuditRecord{
		auditRecord("tok-1", approvals.AuditRequested, base),
		auditRecord("tok-1", approvals.AuditExecuted, base.Add(time.Minute)),
		auditRecord("tok-2", approvals.AuditRequested, base.Add(2*time.Minute)),
		auditRecord("tok-2", approvals.AuditDenied, base.Add(3*time.Minute)),
		auditRecord("user", approvals.AuditExecuted, base.Add(4*time.Minute)),
	}
	for i, rec := range sequence {
		if err := store.Append(t.Context(), rec); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	recent, err := store.Recent(t.Context(), 3)
	if err != nil || len(recent) != 3 {
		t.Fatalf("recent window unexpected: %d (%v)", len(recent), err)
	}
	if recent[0].Correlation != "tok-2" || recent[2].Correlation != "user" {
		t.Fatalf("recent order unexpected: %+v", recent)
	}

	activity, err := store.Activity(t.Context(), 10)
	if err != nil || len(activity) != 3 {
		t.Fatalf("activity view unexpected: %d (%v)", len(activity), err)
	}
	if activity[0].Correlation != "user" || activity[1].Correlation != "tok-2" ||
		activity[2].Correlation != "tok-1" {
		t.Fatalf("activity order unexpected: %+v", activity)
	}
	if activity[1].Outcome != approvals.AuditDenied || activity[1].Events != 2 {
		t.Fatalf("latest outcome per interaction wrong: %+v", activity[1])
	}
	if activity[2].Outcome != approvals.AuditExecuted || activity[2].Events != 2 {
		t.Fatalf("tok-1 should show executed with 2 events: %+v", activity[2])
	}
	if paged, _ := store.Activity(t.Context(), 1); len(paged) != 1 {
		t.Fatalf("activity pagination ignored: %+v", paged)
	}
}

func TestAuditStorageRefusesNonRedactedRecords(t *testing.T) {
	store := abre(t)
	base := time.Unix(1_700_000_000, 0).UTC()
	bad := []approvals.AuditRecord{
		{Correlation: "tok-1", Tool: "copy-export",
			ArgsHash: `C:\Users\me\collection.json`, Outcome: approvals.AuditRequested, At: base},
		{Correlation: "tok-1", Tool: `{"payload":1}`,
			ArgsHash: "abcd", Outcome: approvals.AuditRequested, At: base},
		{Correlation: "home/me/run", Tool: "copy-export",
			ArgsHash: "abcd", Outcome: approvals.AuditRequested, At: base},
	}
	for i, rec := range bad {
		if err := store.Append(t.Context(), rec); err == nil {
			t.Errorf("record %d should be refused as non-redacted", i)
		}
	}
	if recent, _ := store.Recent(t.Context(), 10); len(recent) != 0 {
		t.Fatalf("refused records must not be stored: %+v", recent)
	}
}
