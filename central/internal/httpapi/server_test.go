package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func testConfig() Config {
	return Config{Addr: "127.0.0.1:0", ReadHeader: time.Second,
		Read: 5 * time.Second, Write: 5 * time.Second, Idle: time.Second,
		Shutdown: 2 * time.Second, MaxBodyBytes: 1024, MaxConcurrent: 2}
}

// startServer runs the server for one test and returns its base URL,
// a cancel func and the channel carrying Run's result.
func startServer(t *testing.T, h http.Handler) (string, func(), chan error) {
	t.Helper()
	srv, err := New(testConfig(), h)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	stopped := make(chan struct{})
	go func() { done <- srv.Run(ctx); close(stopped) }()
	t.Cleanup(func() { cancel(); <-stopped })
	return "http://" + srv.Addr(), cancel, done
}

func TestServerRefusesIncompleteBudgets(t *testing.T) {
	if _, err := New(Config{Addr: "127.0.0.1:0"}, http.NotFoundHandler()); err == nil {
		t.Fatal("incomplete server budgets must be refused")
	}
}

// TestServerDrainsInFlightRequestsOnShutdown proves cancellation stops
// new work but lets the request already inside the handler finish.
func TestServerDrainsInFlightRequestsOnShutdown(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	base, cancel, done := startServer(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			entered <- struct{}{}
			<-release
			w.WriteHeader(http.StatusOK)
		}))
	status := make(chan int, 1)
	go func() {
		resp, err := http.Get(base + "/")
		if err != nil {
			status <- 0
			return
		}
		resp.Body.Close()
		status <- resp.StatusCode
	}()
	<-entered // the request is inside the handler
	cancel()
	time.Sleep(100 * time.Millisecond) // let Shutdown begin draining
	close(release)
	if code := <-status; code != http.StatusOK {
		t.Fatalf("in-flight request must drain to 200, got %d", code)
	}
	if err := <-done; err != nil {
		t.Fatalf("run must return nil after a drained shutdown: %v", err)
	}
}
