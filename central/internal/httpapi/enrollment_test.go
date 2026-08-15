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

// fakeStore records lifecycle calls and simulates one installation.
type fakeStore struct {
	enrolled  map[string]string // id -> token hash
	deletions []string
	revoked   []string
	rotations [][3]string
}

func newFakeStore() *fakeStore {
	return &fakeStore{enrolled: map[string]string{}}
}

func (s *fakeStore) EnrollWithConsent(_ context.Context, id, tokenHash,
	deletionHash, purpose string, version int) error {
	if purpose == "" || version < 1 || tokenHash == deletionHash {
		return errors.New("invalid enrollment")
	}
	s.enrolled[id] = tokenHash
	return nil
}

func (s *fakeStore) RotateToken(_ context.Context, id, oldHash,
	newHash string) error {
	if s.enrolled[id] != oldHash {
		return errors.New("stale credential")
	}
	s.enrolled[id] = newHash
	s.rotations = append(s.rotations, [3]string{id, oldHash, newHash})
	return nil
}

func (s *fakeStore) Revoke(_ context.Context, id string) error {
	s.revoked = append(s.revoked, id)
	return nil
}

func (s *fakeStore) RequestDeletion(_ context.Context, id,
	deletionHash string) error {
	if id != "inst-known" || deletionHash != HashToken("right-secret") {
		return errors.New("unknown installation or secret")
	}
	s.deletions = append(s.deletions, id)
	return nil
}

func TestEnrollReturnsOnceShownCredentialsAndStoresHashes(t *testing.T) {
	store := newFakeStore()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(`{"purpose":"product","version":1}`))
	EnrollHandler(store).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("enroll must yield 201: %d %s", rec.Code, rec.Body.String())
	}
	var resp enrollResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response must be json: %v", err)
	}
	if resp.InstallationID == "" || resp.Token == "" || resp.DeletionSecret == "" {
		t.Fatalf("credentials must be returned once: %+v", resp)
	}
	if store.enrolled[resp.InstallationID] != HashToken(resp.Token) {
		t.Fatal("only the token hash may be stored")
	}
}

func TestEnrollRefusesIncompleteConsent(t *testing.T) {
	for _, raw := range []string{`{"purpose":"","version":1}`,
		`{"purpose":"product","version":0}`, `{"purpose":"p","x":1}`} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
		EnrollHandler(newFakeStore()).ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s must yield 400, got %d", raw, rec.Code)
		}
	}
}
