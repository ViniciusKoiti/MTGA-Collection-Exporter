package postgres

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestDeletionCompletesWithTombstoneAndSurvivesRestore: completion
// purges telemetry, blanks secrets, stamps non-identifying evidence,
// and reapplying tombstones kills data restored from a backup.
func TestDeletionCompletesWithTombstoneAndSurvivesRestore(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	repo := InstallationsRepo{Q: pool}
	if err := repo.Enroll(ctx, Installation{ID: "tomb-1",
		TokenHash: "h-secret", DeletionHash: "d-secret"}); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	batch := TelemetryBatch{ID: "tb-1", InstallationID: "tomb-1", Sequence: 1,
		Events: []TelemetryEvent{{Name: "scan_completed"}}}
	if err := IngestBatch(ctx, pool, batch, time.Hour); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if err := RequestDeletion(ctx, pool, "tomb-1", "d-secret"); err != nil {
		t.Fatalf("request deletion: %v", err)
	}
	proof, err := CompleteDeletion(ctx, pool, "tomb-1")
	if err != nil || proof == "" || strings.Contains(proof, "tomb-1") {
		t.Fatalf("proof must exist and identify nothing: %q %v", proof, err)
	}
	var batches int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM telemetry_batches
		WHERE installation_id = 'tomb-1'`).Scan(&batches); err != nil ||
		batches != 0 {
		t.Fatalf("telemetry must be purged: %d %v", batches, err)
	}
	inst, err := repo.ByID(ctx, "tomb-1")
	if err != nil || inst.TokenHash != "" || !inst.Revoked {
		t.Fatalf("secret material must be blanked: %+v %v", inst, err)
	}
	// A backup restore resurrects one purged batch...
	restored := TelemetryBatch{ID: "tb-ghost", InstallationID: "tomb-1",
		Sequence: 9, Events: []TelemetryEvent{{Name: "scan_completed"}}}
	if err := IngestBatch(ctx, pool, restored, time.Hour); err != nil {
		t.Fatalf("simulated restore: %v", err)
	}
	applied, err := ReapplyTombstones(ctx, pool)
	if err != nil || applied != 1 {
		t.Fatalf("tombstone reapplication: %d %v", applied, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM telemetry_batches
		WHERE installation_id = 'tomb-1'`).Scan(&batches); err != nil ||
		batches != 0 {
		t.Fatalf("restored data must die again: %d %v", batches, err)
	}
}
