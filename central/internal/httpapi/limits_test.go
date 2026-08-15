package httpapi

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"
)

// readingHandler drains the body so the body budget can trip.
func readingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestBodyBudgetRefusesOversizedRequests(t *testing.T) {
	base, _, _ := startServer(t, readingHandler())
	resp, err := http.Post(base+"/", "text/plain",
		strings.NewReader(strings.Repeat("x", 4096)))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body must yield 413, got %d", resp.StatusCode)
	}
}

// TestGzipBombIsRefused: a tiny compressed body inflating far past the
// budget must be refused before the handler ever sees it.
func TestGzipBombIsRefused(t *testing.T) {
	base, _, _ := startServer(t, readingHandler())
	var bomb bytes.Buffer
	zw := gzip.NewWriter(&bomb)
	if _, err := zw.Write(bytes.Repeat([]byte{0}, 64*1024)); err != nil {
		t.Fatalf("compress: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, base+"/", &bomb)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("gzip bomb must yield 413, got %d", resp.StatusCode)
	}
}

// TestConcurrencyBudgetRefusesExcessRequests fills every slot and
// proves the next request is refused instead of queued.
func TestConcurrencyBudgetRefusesExcessRequests(t *testing.T) {
	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	base, _, _ := startServer(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			entered <- struct{}{}
			<-release
			w.WriteHeader(http.StatusOK)
		}))
	defer close(release)
	for range 2 { // MaxConcurrent in testConfig
		go func() {
			resp, err := http.Get(base + "/")
			if err == nil {
				resp.Body.Close()
			}
		}()
	}
	<-entered
	<-entered // both slots are now held inside the handler
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("request beyond capacity must yield 503, got %d",
			resp.StatusCode)
	}
}
