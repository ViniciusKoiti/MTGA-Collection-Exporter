package httpapi

import (
	"errors"
	"net/http"
)

// TelemetryHandler ingests one batch: the installation scope comes
// from the principal (never the body), validation is atomic, and a
// replay is acknowledged without persisting anything new.
func TelemetryHandler(policy EventPolicy, store TelemetryStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFrom(r.Context())
		if !ok {
			WriteError(w, r, http.StatusUnauthorized, "unauthenticated",
				"missing principal")
			return
		}
		var req telemetryRequest
		if err := DecodeStrict(r, &req); err != nil ||
			req.BatchID == "" || req.Sequence < 1 {
			WriteError(w, r, http.StatusBadRequest, "invalid_request",
				"batch_id, sequence and events are required")
			return
		}
		if err := policy.validate(req.Events); err != nil {
			WriteError(w, r, http.StatusUnprocessableEntity, "invalid_batch",
				"batch refused by the event policy")
			return
		}
		events := make([]TelemetryEventInput, len(req.Events))
		for i, event := range req.Events {
			events[i] = TelemetryEventInput{Name: event.Name, Attrs: event.Attrs}
		}
		err := store.IngestBatch(r.Context(), principal.InstallationID,
			req.BatchID, req.Sequence, events)
		if errors.Is(err, ErrDuplicateBatch) {
			WriteJSON(w, http.StatusOK,
				map[string]bool{"accepted": true, "duplicate": true})
			return
		}
		if err != nil {
			WriteError(w, r, http.StatusInternalServerError, "internal",
				"ingestion failed")
			return
		}
		WriteJSON(w, http.StatusAccepted, map[string]bool{"accepted": true})
	})
}

// TelemetryRoutes mounts ingestion behind authentication, the
// installation scope and rate-based admission control.
func TelemetryRoutes(policy EventPolicy, store TelemetryStore,
	verify Verifier, rate *RatePolicy, observe func(Observation)) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /v1/telemetry", Chain(TelemetryHandler(policy, store),
		WithRequestID, WithRecovery, WithObserver("/v1/telemetry", observe),
		WithAuth(verify), RequireScope("installation"), rate.WithRate))
	return mux
}
