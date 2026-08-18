package scryfall

import (
	"net/http"
	"testing"
	"time"
)

// TestOfflineFallbackServesLastGoodCacheAsStale: a warm cache older than
// the freshness window plus a dead network serves the cache flagged
// stale — the catalog never invents data and never fails silently.
func TestOfflineFallbackServesLastGoodCacheAsStale(t *testing.T) {
	dir := t.TempDir()
	ts := server(t, nil)
	warm := time.Unix(1_700_000_000, 0)
	if _, err := Open(t.Context(), client(ts.URL, dir, 1), warm, time.Hour); err != nil {
		t.Fatalf("warm-up: %v", err)
	}
	ts.Close() // network gone

	later := warm.Add(48 * time.Hour) // beyond freshness: refresh is forced
	offline := &Client{BaseURL: ts.URL, Dir: dir, Retries: 2,
		HTTP: &http.Client{Timeout: time.Second}}
	catalog, err := Open(t.Context(), offline, later, time.Hour)
	if err != nil {
		t.Fatalf("offline fallback should serve the cache: %v", err)
	}
	if !catalog.Stale || catalog.Meta.Fresh(later, time.Hour) {
		t.Fatalf("fallback must be flagged stale: %+v", catalog.Meta)
	}
	if _, found, _ := catalog.ResolvePorArena(t.Context(), 82183); !found {
		t.Fatal("stale cache still resolves the last good data")
	}
}

// TestNoCacheNoNetworkIsAHardError: nothing to serve means a loud error.
func TestNoCacheNoNetworkIsAHardError(t *testing.T) {
	dead := &Client{BaseURL: "http://127.0.0.1:1", Dir: t.TempDir(), Retries: 1,
		HTTP: &http.Client{Timeout: time.Second}}
	if _, err := Open(t.Context(), dead, time.Unix(1_700_000_000, 0), time.Hour); err == nil {
		t.Fatal("no cache and no network must fail loudly")
	}
}
