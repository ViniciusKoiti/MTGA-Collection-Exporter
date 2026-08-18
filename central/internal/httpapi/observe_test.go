package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestObserverRecordsRouteAndStatus: one low-cardinality record per
// request carrying the registered pattern, not the raw URL.
func TestObserverRecordsRouteAndStatus(t *testing.T) {
	var got []Observation
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), WithObserver("/v1/manifest", func(o Observation) {
		got = append(got, o)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/manifest?x=1", nil))
	if len(got) != 1 {
		t.Fatalf("expected one observation, got %d", len(got))
	}
	if got[0].Route != "/v1/manifest" || got[0].Status != http.StatusTeapot {
		t.Fatalf("observation must carry pattern and status: %+v", got[0])
	}
	if got[0].Duration < 0 {
		t.Fatalf("duration must be measured: %+v", got[0])
	}
}
