package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestTelemetryIngestionIsAtomicAndIdempotent: a replayed sequence or
// batch ID adds nothing, and a failed batch leaves zero events behind.
func TestTelemetryIngestionIsAtomicAndIdempotent(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	repo := InstallationsRepo{Q: pool}
	if err := repo.Enroll(ctx, Installation{ID: "tel-1", TokenHash: "h",
		DeletionHash: "d"}); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	batch := TelemetryBatch{ID: "b-1", InstallationID: "tel-1", Sequence: 1,
		Events: []TelemetryEvent{
			{Name: "scan_completed", Attrs: map[string]string{"result": "ok"}},
			{Name: "export_written", Attrs: map[string]string{"format": "json"}},
		}}
	if err := IngestBatch(ctx, pool, batch, time.Hour); err != nil {
		t.Fatalf("first ingest must succeed: %v", err)
	}
	replay := batch
	replay.ID = "b-2" // new ID, same (installation, sequence)
	if err := IngestBatch(ctx, pool, replay, time.Hour); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("replayed sequence must map to ErrDuplicate: %v", err)
	}
	sameID := batch
	sameID.Sequence = 2 // same batch ID, new sequence
	if err := IngestBatch(ctx, pool, sameID, time.Hour); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("replayed batch id must map to ErrDuplicate: %v", err)
	}
	var events int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM accepted_events`).Scan(&events); err != nil {
		t.Fatalf("count: %v", err)
	}
	if events != 2 {
		t.Fatalf("failed replays must add zero events, got %d", events)
	}
}
