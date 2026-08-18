package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"testing"
	"time"
)

// implementedRoutes is the handwritten route inventory: adding an
// endpoint means updating this list AND central/api/openapi.yaml.
var implementedRoutes = []string{
	"/livez", "/ops/v1/audit", "/ops/v1/catalog", "/ops/v1/deletions",
	"/ops/v1/jobs/{id}", "/ops/v1/status", "/readyz", "/v1/compatibility",
	"/v1/installations", "/v1/installations/deletion",
	"/v1/installations/revoke", "/v1/installations/rotate",
	"/v1/manifest/current", "/v1/telemetry",
}

// TestOpenAPIContractMatchesTheRouteInventory: the contract and the
// inventory must list exactly the same paths, both directions.
func TestOpenAPIContractMatchesTheRouteInventory(t *testing.T) {
	raw, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("contract: %v", err)
	}
	pathLine := regexp.MustCompile(`(?m)^  (/[^:]+):`)
	var contract []string
	for _, match := range pathLine.FindAllStringSubmatch(string(raw), -1) {
		contract = append(contract, match[1])
	}
	slices.Sort(contract)
	inventory := slices.Clone(implementedRoutes)
	slices.Sort(inventory)
	if !slices.Equal(contract, inventory) {
		t.Fatalf("contract and code diverged:\nyaml: %v\ncode: %v",
			contract, inventory)
	}
}

// TestEveryInventoryRouteIsActuallyMounted: each route resolves on one
// of the real routers — anything 404 on all of them is vaporware.
func TestEveryInventoryRouteIsActuallyMounted(t *testing.T) {
	doc := &DocHandler{Source: func(context.Context) (Document, error) {
		return Document{Body: []byte(`{}`), ContentType: "application/json"}, nil
	}, TTL: time.Minute, MaxAge: 60}
	compat := &DocHandler{Source: doc.Source, TTL: time.Minute, MaxAge: 60}
	verify := func(context.Context, string) (Principal, error) {
		return Principal{InstallationID: "i", Scope: "installation"}, nil
	}
	rate := &RatePolicy{Limit: 100, Window: time.Minute, MaxKeys: 8}
	health := http.NewServeMux()
	health.Handle("GET /livez", LivenessHandler())
	ready := readiness(t, nil)
	health.Handle("GET /readyz", ready.Handler(time.Second))
	routers := []http.Handler{
		PublicRoutes(doc, compat, func(Observation) {}),
		InstallationRoutes(newFakeStore(), verify, func(Observation) {}),
		TelemetryRoutes(EventPolicy{Names: map[string]bool{"x": true},
			MaxEvents: 1, MaxAttrs: 1, MaxValueLen: 8},
			&fakeTelemetry{}, verify, rate, func(Observation) {}),
		OpsRoutes(&fakeOps{}, verify, func(Observation) {}),
		health,
	}
	for _, route := range implementedRoutes {
		path := route
		if route == "/ops/v1/jobs/{id}" {
			path = "/ops/v1/jobs/j-1"
		}
		mounted := false
		for _, router := range routers {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusNotFound {
				mounted = true
				break
			}
		}
		if !mounted {
			t.Fatalf("route %s is in the contract but mounted nowhere", route)
		}
	}
}
