package httpapi

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestFeatureGateFlipsWithoutRedeploy: while the flag is off every
// request answers the stable feature_disabled envelope; flipping it
// on serves immediately — and flipping back off is the rollback.
func TestFeatureGateFlipsWithoutRedeploy(t *testing.T) {
	var enabled atomic.Bool
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRequestID, WithFeature(enabled.Load))

	dark := httptest.NewRecorder()
	h.ServeHTTP(dark, httptest.NewRequest(http.MethodPost, "/v1/telemetry", nil))
	if dark.Code != http.StatusServiceUnavailable {
		t.Fatalf("a dark feature must refuse with 503, got %d", dark.Code)
	}
	enabled.Store(true)
	lit := httptest.NewRecorder()
	h.ServeHTTP(lit, httptest.NewRequest(http.MethodPost, "/v1/telemetry", nil))
	if lit.Code != http.StatusOK {
		t.Fatalf("an enabled feature must serve, got %d", lit.Code)
	}
	enabled.Store(false) // the rollback
	back := httptest.NewRecorder()
	h.ServeHTTP(back, httptest.NewRequest(http.MethodPost, "/v1/telemetry", nil))
	if back.Code != http.StatusServiceUnavailable {
		t.Fatalf("the rollback must take effect immediately, got %d", back.Code)
	}
}
