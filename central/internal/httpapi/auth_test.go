package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func authedHandler(t *testing.T, verify Verifier,
	scope string) http.Handler {
	t.Helper()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := PrincipalFrom(r.Context())
		if !ok {
			t.Error("handler must see the principal")
		}
		w.Header().Set("X-Installation", p.InstallationID)
	})
	return Chain(inner, WithRequestID, WithAuth(verify), RequireScope(scope))
}

func TestAuthResolvesHashedTokensOnly(t *testing.T) {
	sum := sha256.Sum256([]byte("secret-token"))
	wantHash := hex.EncodeToString(sum[:])
	verify := func(_ context.Context, hash string) (Principal, error) {
		if hash != wantHash {
			return Principal{}, errors.New("unknown token")
		}
		return Principal{InstallationID: "inst-1", Scope: "installation"}, nil
	}
	h := authedHandler(t, verify, "installation")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("X-Installation") != "inst-1" {
		t.Fatalf("valid token must authenticate: %d", rec.Code)
	}
}

func TestAuthRefusesMissingAndUnknownTokens(t *testing.T) {
	verify := func(context.Context, string) (Principal, error) {
		return Principal{}, errors.New("unknown token")
	}
	h := authedHandler(t, verify, "installation")
	for name, header := range map[string]string{
		"missing": "", "unknown": "Bearer wrong", "malformed": "Basic abc"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s token must yield 401, got %d", name, rec.Code)
		}
	}
}

func TestScopeMismatchIsForbidden(t *testing.T) {
	verify := func(context.Context, string) (Principal, error) {
		return Principal{InstallationID: "inst-1", Scope: "installation"}, nil
	}
	h := authedHandler(t, verify, "operations")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer any")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("scope mismatch must yield 403, got %d", rec.Code)
	}
}
