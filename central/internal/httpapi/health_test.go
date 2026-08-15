package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func readyProbe(context.Context) error { return nil }

func readiness(t *testing.T, overrides map[string]Probe) Readiness {
	t.Helper()
	pick := func(name string) Probe {
		if p, ok := overrides[name]; ok {
			return p
		}
		return readyProbe
	}
	r, err := NewReadiness(pick("schema"), pick("storage"), pick("pool"),
		pick("queue"), pick("signing"), pick("publication"))
	if err != nil {
		t.Fatalf("readiness: %v", err)
	}
	return r
}

func TestReadinessDemandsEveryCapability(t *testing.T) {
	if _, err := NewReadiness(readyProbe, nil, readyProbe, readyProbe,
		readyProbe, readyProbe); err == nil {
		t.Fatal("a missing probe must be refused")
	}
}

func TestReadinessNamesFailingCapabilitiesOnly(t *testing.T) {
	boom := errors.New("connection storm details that must not leak")
	h := readiness(t, map[string]Probe{
		"queue":   func(context.Context) error { return boom },
		"signing": func(context.Context) error { return boom },
	}).Handler(time.Second)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("failing probes must yield 503, got %d", rec.Code)
	}
	var payload struct {
		Ready   bool     `json:"ready"`
		Failing []string `json:"failing"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload.Ready || len(payload.Failing) != 2 ||
		payload.Failing[0] != "queue" || payload.Failing[1] != "signing" {
		t.Fatalf("names only, sorted: %+v", payload)
	}
	if strings.Contains(rec.Body.String(), "storm") {
		t.Fatalf("error text must not leak: %s", rec.Body.String())
	}
}

func TestSlowProbesFailReadinessButNotLiveness(t *testing.T) {
	hang := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}
	ready := readiness(t, map[string]Probe{"storage": hang}).
		Handler(20 * time.Millisecond)
	rec := httptest.NewRecorder()
	ready.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("a hanging probe must fail readiness: %d", rec.Code)
	}
	live := httptest.NewRecorder()
	LivenessHandler().ServeHTTP(live,
		httptest.NewRequest(http.MethodGet, "/livez", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness must stay 200 while unready: %d", live.Code)
	}
}
