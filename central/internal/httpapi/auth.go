package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// Principal identifies an authenticated installation and its scope.
type Principal struct {
	InstallationID string
	Scope          string
}

const principalKey ctxKey = 2

// PrincipalFrom exposes the authenticated principal to handlers.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// Verifier resolves a SHA-256 token hash to a principal; raw tokens
// never reach storage or logs.
type Verifier func(ctx context.Context, tokenHash string) (Principal, error)

// WithAuth authenticates Bearer tokens by hash lookup and refuses
// requests without a resolvable principal.
func WithAuth(verify Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || raw == "" {
				WriteError(w, r, http.StatusUnauthorized, "unauthenticated",
					"missing bearer token")
				return
			}
			sum := sha256.Sum256([]byte(raw))
			p, err := verify(r.Context(), hex.EncodeToString(sum[:]))
			if err != nil {
				WriteError(w, r, http.StatusUnauthorized, "unauthenticated",
					"unknown or revoked token")
				return
			}
			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), principalKey, p)))
		})
	}
}

// RequireScope refuses principals whose scope does not match; the
// operations plane and the installation plane never share handlers.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFrom(r.Context())
			if !ok || p.Scope != scope {
				WriteError(w, r, http.StatusForbidden, "forbidden",
					"scope not permitted")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
