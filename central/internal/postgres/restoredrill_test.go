package postgres

import (
	"context"
	"testing"
	"time"
)

// TestEncryptedBackupRestoreDrill is the isolated restore drill of
// task 7.6: dump the source database, prove the backup encrypts and
// decrypts, replay it into a FRESH container, verify the migrator
// accepts the restored schema, reapply tombstones so purged data
// cannot rise from the backup, and smoke-check the surviving rows.
// The whole drill is timed as the RTO evidence.
func TestEncryptedBackupRestoreDrill(t *testing.T) {
	start := time.Now()
	ctx := context.Background()
	source, sourceDSN := drillContainer(t)
	if err := RunMigrator(ctx, sourceDSN); err != nil {
		t.Fatalf("source migrate: %v", err)
	}
	pool := poolFor(t, sourceDSN)
	repo := InstallationsRepo{Q: pool}
	for _, id := range []string{"drill-alive", "drill-dead"} {
		if err := repo.Enroll(ctx, Installation{ID: id, TokenHash: "h",
			DeletionHash: "d"}); err != nil {
			t.Fatalf("enroll %s: %v", id, err)
		}
		if err := IngestBatch(ctx, pool, TelemetryBatch{ID: "b-" + id,
			InstallationID: id, Sequence: 1, Events: []TelemetryEvent{
				{Name: "scan_completed"}}}, time.Hour); err != nil {
			t.Fatalf("ingest %s: %v", id, err)
		}
	}
	if err := RequestDeletion(ctx, pool, "drill-dead", "d"); err != nil {
		t.Fatalf("request deletion: %v", err)
	}
	if _, err := CompleteDeletion(ctx, pool, "drill-dead"); err != nil {
		t.Fatalf("complete deletion: %v", err)
	}

	backup := encryptRoundTrip(t, dumpDatabase(t, source))

	target, targetDSN := drillContainer(t)
	// Documented restore ordering: the migrator builds schema and
	// roles first; the backup only ever replays data.
	if err := RunMigrator(ctx, targetDSN); err != nil {
		t.Fatalf("target migrate: %v", err)
	}
	restoreDatabase(t, target, backup)
	// The migrator must still accept the restored database as its own.
	if err := RunMigrator(ctx, targetDSN); err != nil {
		t.Fatalf("restored database must satisfy the migrator: %v", err)
	}
	targetPool := poolFor(t, targetDSN)
	// A remnant from an older backup resurrects purged data...
	if err := IngestBatch(ctx, targetPool, TelemetryBatch{ID: "b-ghost",
		InstallationID: "drill-dead", Sequence: 9,
		Events: []TelemetryEvent{{Name: "scan_completed"}}},
		time.Hour); err != nil {
		t.Fatalf("simulated remnant: %v", err)
	}
	applied, err := ReapplyTombstones(ctx, targetPool)
	if err != nil || applied != 1 {
		t.Fatalf("tombstone reapplication after restore: %d %v", applied, err)
	}
	var deadBatches, aliveBatches int
	if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM telemetry_batches
		WHERE installation_id = 'drill-dead'`).Scan(&deadBatches); err != nil ||
		deadBatches != 0 {
		t.Fatalf("purged data must stay dead after restore: %d %v",
			deadBatches, err)
	}
	if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM telemetry_batches
		WHERE installation_id = 'drill-alive'`).Scan(&aliveBatches); err != nil ||
		aliveBatches != 1 {
		t.Fatalf("smoke: surviving data must restore intact: %d %v",
			aliveBatches, err)
	}
	elapsed := time.Since(start)
	t.Logf("restore drill RTO: %s (budget 3m)", elapsed)
	if elapsed > 3*time.Minute {
		t.Fatalf("drill exceeded the RTO budget: %s", elapsed)
	}
}
