package sqlitestore

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// TestTelemetryOutboxSurvivesRestartAndAcksIdempotently: unacked
// events outlive a process restart, acknowledgement is durable, and
// acking the same seq twice changes nothing (central task 5.2).
func TestTelemetryOutboxSurvivesRestartAndAcksIdempotently(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outbox.db")
	ctx := context.Background()
	at := time.Unix(1_700_000_000, 0)

	first, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	for i, name := range []string{"scan_completed", "export_written",
		"sync_failed"} {
		if err := first.AppendTelemetry(ctx, name, at.Add(time.Duration(i)*time.Second),
			map[string]string{"result": "ok"}); err != nil {
			t.Fatalf("append %s: %v", name, err)
		}
	}
	if err := first.Ack(ctx, 1); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := Open(path) // the "restart"
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	pending, err := reopened.Pending(ctx, 10)
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(pending) != 2 || pending[0].Seq != 2 || pending[1].Seq != 3 {
		t.Fatalf("unacked events must survive the restart in order: %+v",
			pending)
	}
	if pending[0].Name != "export_written" ||
		pending[0].Attrs["result"] != "ok" || !pending[0].At.After(at.Add(-time.Second)) {
		t.Fatalf("event payload must round-trip: %+v", pending[0])
	}
	if err := reopened.Ack(ctx, 3); err != nil {
		t.Fatalf("ack all: %v", err)
	}
	if err := reopened.Ack(ctx, 3); err != nil {
		t.Fatalf("re-ack must be a no-op, not an error: %v", err)
	}
	pending, err = reopened.Pending(ctx, 10)
	if err != nil || len(pending) != 0 {
		t.Fatalf("acked events must never resend: %+v %v", pending, err)
	}
}
