package httpapi

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestClientCancellationReachesTheHandler: dropping the request must
// cancel the handler context so no orphaned work keeps running.
func TestClientCancellationReachesTheHandler(t *testing.T) {
	sawCancel := make(chan struct{})
	base, _, _ := startServer(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
			close(sawCancel)
		}))
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()
	time.Sleep(100 * time.Millisecond) // let the request enter the handler
	cancel()
	select {
	case <-sawCancel:
	case <-time.After(3 * time.Second):
		t.Fatal("client cancellation must reach the handler context")
	}
}

// TestHeaderTimeoutClosesSilentConnections: a connection that never
// sends headers is closed within the read-header budget.
func TestHeaderTimeoutClosesSilentConnections(t *testing.T) {
	base, _, _ := startServer(t, http.NotFoundHandler())
	conn, err := net.Dial("tcp", strings.TrimPrefix(base, "http://"))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("deadline: %v", err)
	}
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err == nil {
		t.Fatal("server must close the silent connection")
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("connection outlived the read-header budget")
	}
}
