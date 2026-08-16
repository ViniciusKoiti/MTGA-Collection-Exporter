package scryfall

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// server fakes the Scryfall bulk endpoints; failures counts how many
// initial requests answer 500 before succeeding.
func server(t *testing.T, failures *atomic.Int32) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	fail := func(w http.ResponseWriter) bool {
		if failures != nil && failures.Add(-1) >= 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return true
		}
		return false
	}
	var ts *httptest.Server
	mux.HandleFunc("/bulk-data", func(w http.ResponseWriter, _ *http.Request) {
		if fail(w) {
			return
		}
		fmt.Fprintf(w, `{"data":[{"type":"oracle_cards","updated_at":"x","download_uri":"%s/oracle"},
			{"type":"default_cards","updated_at":"2026-08-15T10:00:00Z","download_uri":"%s/cards"}]}`,
			ts.URL, ts.URL)
	})
	mux.HandleFunc("/cards", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[
			{"id":"p1","oracle_id":"o1","name":"Lightning Strike","set":"dmu",
			 "collector_number":"137","arena_id":82183,
			 "colors":["R"],"cmc":2.0,"type_line":"Instant",
			 "legalities":{"standard":"legal","modern":"legal","pauper":"not_legal"}},
			{"id":"p2","oracle_id":"o2","name":"Paper Only Card","set":"xxx",
			 "collector_number":"1","arena_id":0}
		]`)
	})
	ts = httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func client(url, dir string, retries int) *Client {
	return &Client{BaseURL: url, Dir: dir, Retries: retries,
		HTTP: &http.Client{Timeout: 2 * time.Second}}
}

func TestFreshFetchCachesVersionAndResolves(t *testing.T) {
	ts := server(t, nil)
	now := time.Unix(1_700_000_000, 0)
	catalog, err := Open(t.Context(), client(ts.URL, t.TempDir(), 1), now, time.Hour)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if catalog.Stale || catalog.Meta.Version != "2026-08-15T10:00:00Z" ||
		!catalog.Meta.Fresh(now, time.Hour) {
		t.Fatalf("metadata unexpected: %+v", catalog.Meta)
	}
	identity, found, _ := catalog.ResolvePorArena(t.Context(), 82183)
	if !found || identity.Name != "Lightning Strike" || identity.Printing != "p1" {
		t.Fatalf("arena resolution failed: %+v", identity)
	}
	if _, found, _ := catalog.ResolvePorArena(t.Context(), 0); found {
		t.Fatal("cards without arena id must not enter the arena index")
	}
	if _, found, _ := catalog.ResolvePorNome(t.Context(), "paper only card"); !found {
		t.Fatal("name resolution should be case-insensitive")
	}
}

func TestBoundedRetriesRecoverFromTransientFailures(t *testing.T) {
	var failures atomic.Int32
	failures.Store(2)
	ts := server(t, &failures)
	catalog, err := Open(t.Context(), client(ts.URL, t.TempDir(), 3),
		time.Unix(1_700_000_000, 0), time.Hour)
	if err != nil || catalog.Stale {
		t.Fatalf("third attempt should succeed inside the budget: %v", err)
	}
	var tooMany atomic.Int32
	tooMany.Store(10)
	if _, err := Open(t.Context(), client(server(t, &tooMany).URL, t.TempDir(), 2),
		time.Unix(1_700_000_000, 0), time.Hour); err == nil {
		t.Fatal("exhausted retry budget without cache must fail")
	}
}
