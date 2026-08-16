package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

type recordingTelemetry struct {
	events []TelemetryEventInput
}

func (s *recordingTelemetry) IngestBatch(_ context.Context, _, _ string,
	_ int64, events []TelemetryEventInput) error {
	s.events = append(s.events, events...)
	return nil
}

// TestAcceptedBatchGoldenPinsThePrivacySurface: the golden file IS
// the privacy contract — exactly which fields an accepted event may
// carry into storage. Any new field added to the wire or the input
// type changes this serialization and fails here first.
func TestAcceptedBatchGoldenPinsThePrivacySurface(t *testing.T) {
	store := &recordingTelemetry{}
	policy := EventPolicy{Names: map[string]bool{"scan_completed": true,
		"export_written": true}, MaxEvents: 4, MaxAttrs: 2, MaxValueLen: 32}
	verify := func(context.Context, string) (Principal, error) {
		return Principal{InstallationID: "inst-1", Scope: "installation"}, nil
	}
	rate := &RatePolicy{Limit: 5, Window: time.Minute, MaxKeys: 4}
	rig := TelemetryRoutes(policy, store, verify, rate, func(Observation) {})
	rec := postBatch(rig, `{"batch_id":"b-g","sequence":1,"events":[`+
		`{"name":"scan_completed","attrs":{"result":"ok"}},`+
		`{"name":"export_written","attrs":{"format":"json"}}]}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("batch must be accepted: %d %s", rec.Code, rec.Body.String())
	}
	got, err := json.MarshalIndent(store.events, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want, err := os.ReadFile("testdata/accepted_batch.golden")
	if err != nil {
		t.Fatalf("golden: %v", err)
	}
	if string(got)+"\n" != string(want) {
		t.Fatalf("privacy surface changed:\n--- golden ---\n%s\n--- got ---\n%s",
			want, got)
	}
}
