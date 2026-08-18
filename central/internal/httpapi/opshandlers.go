package httpapi

import (
	"net/http"
	"strconv"
)

// opsListLimit caps list sizes on the operations plane; the cap is a
// server decision, never the caller's.
const opsListLimit = 100

func opsLimit(r *http.Request) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > opsListLimit {
		return opsListLimit
	}
	return limit
}

func opsGet[T any](fetch func(*http.Request) (T, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, err := fetch(r)
		if err != nil {
			WriteError(w, r, http.StatusNotFound, "not_found",
				"unknown resource")
			return
		}
		WriteJSON(w, http.StatusOK, out)
	})
}

// OpsRoutes mounts the separately authenticated, read-only operations
// plane; every handler requires the operations scope and the verifier
// is expected to resolve operator credentials, not installation ones.
func OpsRoutes(source OpsSource, verify Verifier,
	observe func(Observation)) http.Handler {
	guard := func(pattern string, h http.Handler) http.Handler {
		return Chain(h, WithRequestID, WithRecovery,
			WithObserver(pattern, observe), WithAuth(verify),
			RequireScope("operations"))
	}
	mux := http.NewServeMux()
	mux.Handle("GET /ops/v1/status", guard("/ops/v1/status",
		opsGet(func(r *http.Request) (OpsStatus, error) {
			return source.Status(r.Context())
		})))
	mux.Handle("GET /ops/v1/jobs/{id}", guard("/ops/v1/jobs/{id}",
		opsGet(func(r *http.Request) (OpsJob, error) {
			return source.Job(r.Context(), r.PathValue("id"))
		})))
	mux.Handle("GET /ops/v1/deletions", guard("/ops/v1/deletions",
		opsGet(func(r *http.Request) ([]OpsDeletion, error) {
			return source.Deletions(r.Context(), opsLimit(r))
		})))
	mux.Handle("GET /ops/v1/audit", guard("/ops/v1/audit",
		opsGet(func(r *http.Request) ([]OpsAudit, error) {
			return source.Audit(r.Context(), opsLimit(r))
		})))
	mux.Handle("GET /ops/v1/catalog", guard("/ops/v1/catalog",
		opsGet(func(r *http.Request) (OpsCatalog, error) {
			return source.Catalog(r.Context())
		})))
	return mux
}
