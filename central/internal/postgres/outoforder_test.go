package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestOutOfOrderSequencesAreAcceptedOnceEach: batches may arrive out
// of order — each sequence lands exactly once, and a replay under a
// fresh batch ID still maps to the stable duplicate error.
func TestOutOfOrderSequencesAreAcceptedOnceEach(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	repo := InstallationsRepo{Q: pool}
	if err := repo.Enroll(ctx, Installation{ID: "ooo-1", TokenHash: "h",
		DeletionHash: "d"}); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	event := []TelemetryEvent{{Name: "scan_completed"}}
	if err := IngestBatch(ctx, pool, TelemetryBatch{ID: "ooo-b2",
		InstallationID: "ooo-1", Sequence: 2, Events: event},
		time.Hour); err != nil {
		t.Fatalf("later sequence first must be accepted: %v", err)
	}
	if err := IngestBatch(ctx, pool, TelemetryBatch{ID: "ooo-b1",
		InstallationID: "ooo-1", Sequence: 1, Events: event},
		time.Hour); err != nil {
		t.Fatalf("earlier sequence later must be accepted: %v", err)
	}
	replay := IngestBatch(ctx, pool, TelemetryBatch{ID: "ooo-b3",
		InstallationID: "ooo-1", Sequence: 2, Events: event}, time.Hour)
	if !errors.Is(replay, ErrDuplicate) {
		t.Fatalf("replayed sequence must stay a duplicate: %v", replay)
	}
	var batches int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM telemetry_batches
		WHERE installation_id = 'ooo-1'`).Scan(&batches); err != nil ||
		batches != 2 {
		t.Fatalf("exactly two batches must exist: %d %v", batches, err)
	}
}
