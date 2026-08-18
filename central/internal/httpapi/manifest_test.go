package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func docSource(calls *atomic.Int64, delay time.Duration) DocumentSource {
	return func(context.Context) (Document, error) {
		calls.Add(1)
		time.Sleep(delay)
		return Document{Body: []byte(`{"v":1}`),
			ContentType: "application/json"}, nil
	}
}

func TestDocHandlerServesWithETagAndRevalidates(t *testing.T) {
	var calls atomic.Int64
	h := &DocHandler{Source: docSource(&calls, 0),
		TTL: time.Minute, MaxAge: 60, Immutable: true}
	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	etag := first.Header().Get("ETag")
	if first.Code != http.StatusOK || etag == "" ||
		first.Header().Get("Cache-Control") != "public, max-age=60, immutable" {
		t.Fatalf("first fetch must serve with cache headers: %d %q",
			first.Code, first.Header())
	}
	second := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", etag)
	h.ServeHTTP(second, req)
	if second.Code != http.StatusNotModified || second.Body.Len() != 0 {
		t.Fatalf("matching etag must yield an empty 304, got %d", second.Code)
	}
	if calls.Load() != 1 {
		t.Fatalf("cache must absorb the revalidation, got %d calls", calls.Load())
	}
}

// TestDocHandlerCollapsesConcurrentMisses: a cold cache hit by many
// requests at once reaches the source exactly once.
func TestDocHandlerCollapsesConcurrentMisses(t *testing.T) {
	var calls atomic.Int64
	h := &DocHandler{Source: docSource(&calls, 100*time.Millisecond),
		TTL: time.Minute, MaxAge: 60}
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("collapsed miss must still serve: %d", rec.Code)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("singleflight must collapse misses to 1 call, got %d",
			calls.Load())
	}
}

func TestDocHandlerRefusesWritesAndExpires(t *testing.T) {
	var calls atomic.Int64
	h := &DocHandler{Source: docSource(&calls, 0),
		TTL: 10 * time.Millisecond, MaxAge: 1}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("writes must be refused with 405, got %d", rec.Code)
	}
	h.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil))
	time.Sleep(20 * time.Millisecond)
	h.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil))
	if calls.Load() != 2 {
		t.Fatalf("expired cache must refresh once, got %d calls", calls.Load())
	}
}
