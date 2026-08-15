package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// EnrollmentStore is the persistence port of the installation
// lifecycle endpoints; the postgres adapters satisfy it.
type EnrollmentStore interface {
	EnrollWithConsent(ctx context.Context, id, tokenHash, deletionHash,
		purpose string, version int) error
	RotateToken(ctx context.Context, id, oldHash, newHash string) error
	Revoke(ctx context.Context, id string) error
	RequestDeletion(ctx context.Context, id, deletionHash string) error
}

// newSecret returns a fresh random credential in hex.
func newSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// HashToken derives the stored form of a credential; raw values are
// returned to the client exactly once and never persisted.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

type enrollRequest struct {
	Purpose string `json:"purpose"`
	Version int    `json:"version"`
}

type enrollResponse struct {
	InstallationID string `json:"installation_id"`
	Token          string `json:"token"`
	DeletionSecret string `json:"deletion_secret"`
}

// EnrollHandler creates the installation, its consent receipt and its
// once-shown credentials in one store call.
func EnrollHandler(store EnrollmentStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req enrollRequest
		if err := DecodeStrict(r, &req); err != nil ||
			req.Purpose == "" || req.Version < 1 {
			WriteError(w, r, http.StatusBadRequest, "invalid_request",
				"purpose and version are required")
			return
		}
		id, errID := newSecret()
		token, errToken := newSecret()
		deletion, errDeletion := newSecret()
		if errID != nil || errToken != nil || errDeletion != nil {
			WriteError(w, r, http.StatusInternalServerError, "internal",
				"credential source unavailable")
			return
		}
		id = "inst-" + id[:16]
		if err := store.EnrollWithConsent(r.Context(), id, HashToken(token),
			HashToken(deletion), req.Purpose, req.Version); err != nil {
			WriteError(w, r, http.StatusInternalServerError, "internal",
				"enrollment failed")
			return
		}
		WriteJSON(w, http.StatusCreated, enrollResponse{InstallationID: id,
			Token: token, DeletionSecret: deletion})
	})
}
