package centralclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// stubCentral implements the wire contract of the installation API.
func stubCentral(t *testing.T, requests *atomic.Int64) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/installations", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		var body struct {
			Purpose string `json:"purpose"`
			Version int    `json:"version"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil ||
			body.Purpose == "" || body.Version < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Credentials{InstallationID: "inst-9",
			Token: "tok-1", DeletionSecret: "del-1"})
	})
	mux.HandleFunc("POST /v1/installations/rotate", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("Authorization") != "Bearer tok-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "tok-2"})
	})
	mux.HandleFunc("POST /v1/installations/revoke", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/installations/deletion", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		var body struct {
			InstallationID string `json:"installation_id"`
			DeletionSecret string `json:"deletion_secret"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil ||
			body.DeletionSecret != "del-1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}
