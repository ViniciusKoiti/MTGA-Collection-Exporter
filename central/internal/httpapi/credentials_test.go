package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// verifierFor authenticates one raw token against the fake store.
func verifierFor(store *fakeStore, id string) Verifier {
	return func(_ context.Context, hash string) (Principal, error) {
		if store.enrolled[id] != hash {
			return Principal{}, errors.New("unknown token")
		}
		return Principal{InstallationID: id, Scope: "installation"}, nil
	}
}

func TestRotationSwapsThePresentedCredential(t *testing.T) {
	store := newFakeStore()
	store.enrolled["inst-1"] = HashToken("current-token")
	routes := InstallationRoutes(store, verifierFor(store, "inst-1"), func(Observation) {})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/installations/rotate", nil)
	req.Header.Set("Authorization", "Bearer current-token")
	routes.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("rotation must succeed: %d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("rotation response must be json: %v", err)
	}
	if store.enrolled["inst-1"] != HashToken(resp["token"]) {
		t.Fatal("store must hold the hash of the newly issued token")
	}
	replay := httptest.NewRecorder()
	again := httptest.NewRequest(http.MethodPost, "/v1/installations/rotate", nil)
	again.Header.Set("Authorization", "Bearer current-token")
	routes.ServeHTTP(replay, again)
	if replay.Code != http.StatusUnauthorized {
		t.Fatalf("the old token must die with the rotation: %d", replay.Code)
	}
}

func TestRevokeRequiresTheInstallationScope(t *testing.T) {
	store := newFakeStore()
	store.enrolled["inst-1"] = HashToken("tok")
	routes := InstallationRoutes(store, verifierFor(store, "inst-1"), func(Observation) {})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/installations/revoke", nil)
	req.Header.Set("Authorization", "Bearer tok")
	routes.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || len(store.revoked) != 1 {
		t.Fatalf("authenticated revoke must succeed: %d %v",
			rec.Code, store.revoked)
	}
}

func TestDeletionIsGatedByTheSecret(t *testing.T) {
	store := newFakeStore()
	routes := InstallationRoutes(store, verifierFor(store, "none"), func(Observation) {})
	for body, want := range map[string]int{
		`{"installation_id":"inst-known","deletion_secret":"right-secret"}`: http.StatusAccepted,
		`{"installation_id":"inst-known","deletion_secret":"wrong"}`:        http.StatusNotFound,
		`{"installation_id":"inst-other","deletion_secret":"right-secret"}`: http.StatusNotFound,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost,
			"/v1/installations/deletion", strings.NewReader(body))
		routes.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s: expected %d, got %d", body, want, rec.Code)
		}
	}
	if len(store.deletions) != 1 {
		t.Fatalf("exactly one deletion must be filed: %v", store.deletions)
	}
}
