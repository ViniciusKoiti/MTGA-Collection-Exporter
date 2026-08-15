package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// Probe answers whether one capability is ready right now.
type Probe func(ctx context.Context) error

// Readiness aggregates the six capabilities the API needs before it
// may receive traffic. Liveness is deliberately NOT this: a live
// process can be unready, and killing it for unreadiness would turn
// every dependency blip into a crash loop.
type Readiness struct {
	probes map[string]Probe
}

// NewReadiness demands every capability explicitly so none can be
// forgotten by omission.
func NewReadiness(schema, storage, pool, queue, signing,
	publication Probe) (Readiness, error) {
	probes := map[string]Probe{"schema": schema, "storage": storage,
		"pool": pool, "queue": queue, "signing": signing,
		"publication": publication}
	for name, probe := range probes {
		if probe == nil {
			return Readiness{}, fmt.Errorf("httpapi: missing %s probe", name)
		}
	}
	return Readiness{probes: probes}, nil
}

// LivenessHandler answers 200 while the process serves at all.
func LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]bool{"alive": true})
	})
}

// Handler runs every probe under its own timeout and reports failing
// capability names only — no internals, no error text on the wire.
func (h Readiness) Handler(timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failing := make([]string, 0, len(h.probes))
		for name, probe := range h.probes {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			if err := probe(ctx); err != nil {
				failing = append(failing, name)
			}
			cancel()
		}
		sort.Strings(failing)
		if len(failing) > 0 {
			WriteJSON(w, http.StatusServiceUnavailable,
				map[string]any{"ready": false, "failing": failing})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"ready": true})
	})
}
