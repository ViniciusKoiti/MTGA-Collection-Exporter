package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeOps struct{ lastLimit int }

func (f *fakeOps) Status(context.Context) (OpsStatus, error) {
	return OpsStatus{SchemaReady: true, PoolMax: 4}, nil
}

func (f *fakeOps) Job(_ context.Context, id string) (OpsJob, error) {
	if id != "j-1" {
		return OpsJob{}, errors.New("unknown job")
	}
	return OpsJob{ID: "j-1", Kind: "publish", Status: "succeeded"}, nil
}

func (f *fakeOps) Deletions(_ context.Context, limit int) ([]OpsDeletion, error) {
	f.lastLimit = limit
	return []OpsDeletion{}, nil
}

func (f *fakeOps) Audit(_ context.Context, limit int) ([]OpsAudit, error) {
	f.lastLimit = limit
	return []OpsAudit{}, nil
}

func (f *fakeOps) Catalog(context.Context) (OpsCatalog, error) {
	return OpsCatalog{SnapshotID: "snap-1"}, nil
}

func opsRig(ops *fakeOps, scope string) http.Handler {
	verify := func(context.Context, string) (Principal, error) {
		return Principal{InstallationID: "op-1", Scope: scope}, nil
	}
	return OpsRoutes(ops, verify, func(Observation) {})
}

func opsGetPath(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer op-token")
	h.ServeHTTP(rec, req)
	return rec
}

// TestOpsPlaneIsSeparatelyAuthenticated: an installation-scoped
// principal is refused on every operations endpoint.
func TestOpsPlaneIsSeparatelyAuthenticated(t *testing.T) {
	rig := opsRig(&fakeOps{}, "installation")
	for _, path := range []string{"/ops/v1/status", "/ops/v1/jobs/j-1",
		"/ops/v1/deletions", "/ops/v1/audit", "/ops/v1/catalog"} {
		if rec := opsGetPath(rig, path); rec.Code != http.StatusForbidden {
			t.Fatalf("%s: installation scope must be refused, got %d",
				path, rec.Code)
		}
	}
}

func TestOpsEndpointsAnswerFixedQuestions(t *testing.T) {
	ops := &fakeOps{}
	rig := opsRig(ops, "operations")
	if rec := opsGetPath(rig, "/ops/v1/status"); rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
	}
	if rec := opsGetPath(rig, "/ops/v1/jobs/j-1"); rec.Code != http.StatusOK {
		t.Fatalf("job: %d", rec.Code)
	}
	if rec := opsGetPath(rig, "/ops/v1/jobs/j-404"); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown job must be 404, got %d", rec.Code)
	}
	if rec := opsGetPath(rig, "/ops/v1/audit?limit=100000"); rec.Code != http.StatusOK {
		t.Fatalf("audit: %d", rec.Code)
	}
	if ops.lastLimit != opsListLimit {
		t.Fatalf("list limits must be capped server-side, got %d", ops.lastLimit)
	}
	if rec := opsGetPath(rig, "/ops/v1/deletions?limit=5"); rec.Code != http.StatusOK {
		t.Fatalf("deletions: %d", rec.Code)
	}
	if ops.lastLimit != 5 {
		t.Fatalf("valid limits must pass through, got %d", ops.lastLimit)
	}
}
