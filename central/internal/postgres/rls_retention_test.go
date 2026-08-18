package postgres

import (
	"context"
	"database/sql"
	"testing"
)

// TestWorkerRetentionCrossesInstallations: expiry is the worker's job
// and must reach every installation, unlike the scoped telemetry path.
func TestWorkerRetentionCrossesInstallations(t *testing.T) {
	dsn := containerDSN(t)
	ctx := context.Background()
	if err := RunMigrator(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, stmt := range []string{
		`INSERT INTO installations (id, token_hash, deletion_hash) VALUES
			('inst-1','t1','d1'), ('inst-2','t2','d2')`,
		`INSERT INTO telemetry_batches (id, installation_id, sequence) VALUES
			('b-1','inst-1',1), ('b-2','inst-2',1)`,
		`INSERT INTO accepted_events (batch_id, name, expires_at) VALUES
			('b-1','sync_completed', now() - interval '1 hour'),
			('b-2','sync_completed', now() - interval '1 hour'),
			('b-2','export_completed', now() + interval '1 hour')`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE central_worker`); err != nil {
		t.Fatalf("role: %v", err)
	}
	result, err := tx.ExecContext(ctx,
		`DELETE FROM accepted_events WHERE expires_at < now()`)
	if err != nil {
		t.Fatalf("worker retention delete: %v", err)
	}
	deleted, _ := result.RowsAffected()
	if deleted != 2 { // one expired event per installation
		t.Fatalf("retention must cross installations: deleted %d", deleted)
	}
	var remaining int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM accepted_events`).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("only the unexpired event should remain: %d (%v)", remaining, err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}
