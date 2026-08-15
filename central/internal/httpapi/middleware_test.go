package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRequestIDIsServerOwned: the ID is generated per request, exposed
// on the response and available to the handler context.
func TestRequestIDIsServerOwned(t *testing.T) {
	var seen string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}), WithRequestID)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "client-forged")
	h.ServeHTTP(rec, req)
	got := rec.Header().Get("X-Request-Id")
	if got == "" || got == "client-forged" || got != seen {
		t.Fatalf("id must be server-owned and shared: header %q ctx %q",
			got, seen)
	}
}

// TestRecoveryYieldsStableEnvelope: a panicking handler becomes the
// closed-vocabulary error payload with the request ID attached.
func TestRecoveryYieldsStableEnvelope(t *testing.T) {
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), WithRequestID, WithRecovery)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic must yield 500, got %d", rec.Code)
	}
	var payload map[string]apiError
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("envelope must be json: %v", err)
	}
	e := payload["error"]
	if e.Code != "internal" || e.RequestID == "" ||
		strings.Contains(e.Message, "boom") {
		t.Fatalf("envelope must be stable and leak nothing: %+v", e)
	}
}
