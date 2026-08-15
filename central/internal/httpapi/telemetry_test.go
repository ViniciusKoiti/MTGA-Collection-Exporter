package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeTelemetry struct {
	batches map[string]string // batch id -> installation
	events  int
}

func (s *fakeTelemetry) IngestBatch(_ context.Context, installationID,
	batchID string, _ int64, events []TelemetryEventInput) error {
	if s.batches == nil {
		s.batches = map[string]string{}
	}
	if _, dup := s.batches[batchID]; dup {
		return ErrDuplicateBatch
	}
	s.batches[batchID] = installationID
	s.events += len(events)
	return nil
}

func telemetryRig(store *fakeTelemetry) http.Handler {
	policy := EventPolicy{Names: map[string]bool{"scan_completed": true},
		MaxEvents: 4, MaxAttrs: 2, MaxValueLen: 32}
	verify := func(context.Context, string) (Principal, error) {
		return Principal{InstallationID: "inst-1", Scope: "installation"}, nil
	}
	rate := &RatePolicy{Limit: 3, Window: time.Minute, MaxKeys: 4}
	return TelemetryRoutes(policy, store, verify, rate, func(Observation) {})
}

func postBatch(h http.Handler, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/telemetry",
		strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer tok")
	h.ServeHTTP(rec, req)
	return rec
}

func TestTelemetryScopesBatchesToThePrincipal(t *testing.T) {
	store := &fakeTelemetry{}
	rec := postBatch(telemetryRig(store),
		`{"batch_id":"b-1","sequence":1,"events":[{"name":"scan_completed","attrs":{"result":"ok"}}]}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("valid batch must be accepted: %d %s", rec.Code, rec.Body.String())
	}
	if store.batches["b-1"] != "inst-1" {
		t.Fatalf("scope must come from the principal: %v", store.batches)
	}
	replay := postBatch(telemetryRig(store), // fresh rate window, same store
		`{"batch_id":"b-1","sequence":1,"events":[{"name":"scan_completed","attrs":{}}]}`)
	if replay.Code != http.StatusOK || store.events != 1 {
		t.Fatalf("replay must acknowledge without persisting: %d events %d",
			replay.Code, store.events)
	}
}

func TestTelemetryValidationIsAtomic(t *testing.T) {
	store := &fakeTelemetry{}
	rig := telemetryRig(store)
	bad := `{"batch_id":"b-2","sequence":1,"events":[` +
		`{"name":"scan_completed","attrs":{}},{"name":"raw_log","attrs":{}}]}`
	if rec := postBatch(rig, bad); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("one bad event must refuse the whole batch: %d", rec.Code)
	}
	if store.events != 0 {
		t.Fatalf("refused batch must persist nothing, got %d", store.events)
	}
}

func TestTelemetryAdmissionControlRefusesFloods(t *testing.T) {
	rig := telemetryRig(&fakeTelemetry{})
	body := `{"batch_id":"b-%d","sequence":1,"events":[{"name":"scan_completed","attrs":{}}]}`
	var last int
	for i := range 4 {
		last = postBatch(rig, strings.Replace(body, "%d",
			string(rune('0'+i)), 1)).Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("the flood must hit admission control, got %d", last)
	}
}
