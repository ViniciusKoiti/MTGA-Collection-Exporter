package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicRoutesServeManifestAndCompatibility(t *testing.T) {
	source := func(body string) DocumentSource {
		return func(context.Context) (Document, error) {
			return Document{Body: []byte(body),
				ContentType: "application/json"}, nil
		}
	}
	var seen []Observation
	routes := PublicRoutes(
		&DocHandler{Source: source(`{"manifest":1}`), TTL: time.Minute, MaxAge: 60},
		&DocHandler{Source: source(`{"compat":1}`), TTL: time.Minute, MaxAge: 60},
		func(o Observation) { seen = append(seen, o) })
	for path, want := range map[string]string{
		"/v1/manifest/current": `{"manifest":1}`,
		"/v1/compatibility":    `{"compat":1}`,
	} {
		rec := httptest.NewRecorder()
		routes.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Fatalf("%s: expected %s, got %d %s", path, want,
				rec.Code, rec.Body.String())
		}
		if rec.Header().Get("X-Request-Id") == "" {
			t.Fatalf("%s: public routes must carry a request id", path)
		}
	}
	if len(seen) != 2 || seen[0].Route == seen[1].Route {
		t.Fatalf("each route must observe under its own pattern: %+v", seen)
	}
}
