package httpapi

import "net/http"

// PublicRoutes mounts the unauthenticated read-only endpoints — the
// current manifest and the compatibility matrix — behind the request
// ID, recovery and observer middleware.
func PublicRoutes(manifest, compatibility *DocHandler,
	observe func(Observation)) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/manifest/current", Chain(manifest, WithRequestID,
		WithRecovery, WithObserver("/v1/manifest/current", observe)))
	mux.Handle("/v1/compatibility", Chain(compatibility, WithRequestID,
		WithRecovery, WithObserver("/v1/compatibility", observe)))
	return mux
}
