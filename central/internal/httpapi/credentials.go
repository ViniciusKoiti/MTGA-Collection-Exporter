package httpapi

import "net/http"

// RotateHandler swaps the caller's credential via compare-and-swap on
// the presented token's hash and returns the new token exactly once.
func RotateHandler(store EnrollmentStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := PrincipalFrom(r.Context())
		if !ok {
			WriteError(w, r, http.StatusUnauthorized, "unauthenticated",
				"missing principal")
			return
		}
		token, err := newSecret()
		if err != nil {
			WriteError(w, r, http.StatusInternalServerError, "internal",
				"credential source unavailable")
			return
		}
		if err := store.RotateToken(r.Context(), p.InstallationID,
			p.TokenHash, HashToken(token)); err != nil {
			WriteError(w, r, http.StatusConflict, "conflict",
				"credential no longer current")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"token": token})
	})
}

// RevokeHandler disables the caller's credential immediately.
func RevokeHandler(store EnrollmentStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := PrincipalFrom(r.Context())
		if !ok {
			WriteError(w, r, http.StatusUnauthorized, "unauthenticated",
				"missing principal")
			return
		}
		if err := store.Revoke(r.Context(), p.InstallationID); err != nil {
			WriteError(w, r, http.StatusInternalServerError, "internal",
				"revocation failed")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

type deletionRequest struct {
	InstallationID string `json:"installation_id"`
	DeletionSecret string `json:"deletion_secret"`
}

// DeletionHandler files a deletion request gated by the deletion
// secret; a wrong secret is indistinguishable from an unknown ID.
func DeletionHandler(store EnrollmentStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req deletionRequest
		if err := DecodeStrict(r, &req); err != nil ||
			req.InstallationID == "" || req.DeletionSecret == "" {
			WriteError(w, r, http.StatusBadRequest, "invalid_request",
				"installation_id and deletion_secret are required")
			return
		}
		if err := store.RequestDeletion(r.Context(), req.InstallationID,
			HashToken(req.DeletionSecret)); err != nil {
			WriteError(w, r, http.StatusNotFound, "not_found",
				"unknown installation or secret")
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}
