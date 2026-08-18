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

// InstallationRoutes mounts the installation lifecycle endpoints; the
// credential-mutating ones require the installation scope, while
// enrollment and secret-gated deletion stand on their own proofs.
func InstallationRoutes(store EnrollmentStore, verify Verifier,
	observe func(Observation)) http.Handler {
	public := func(pattern string, h http.Handler) http.Handler {
		return Chain(h, WithRequestID, WithRecovery,
			WithObserver(pattern, observe))
	}
	authed := func(pattern string, h http.Handler) http.Handler {
		return Chain(h, WithRequestID, WithRecovery,
			WithObserver(pattern, observe), WithAuth(verify),
			RequireScope("installation"))
	}
	mux := http.NewServeMux()
	mux.Handle("POST /v1/installations",
		public("/v1/installations", EnrollHandler(store)))
	mux.Handle("POST /v1/installations/rotate",
		authed("/v1/installations/rotate", RotateHandler(store)))
	mux.Handle("POST /v1/installations/revoke",
		authed("/v1/installations/revoke", RevokeHandler(store)))
	mux.Handle("POST /v1/installations/deletion",
		public("/v1/installations/deletion", DeletionHandler(store)))
	return mux
}
