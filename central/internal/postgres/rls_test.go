package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

// asTelemetry runs stmt as central_telemetry scoped to one installation
// inside a rolled-back transaction; scan receives one row when wanted.
func asTelemetry(ctx context.Context, db *sql.DB, installation, stmt string, scan func(*sql.Row) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE central_telemetry`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf(`SET LOCAL app.installation_id = '%s'`, installation)); err != nil {
		return err
	}
	if scan != nil {
		return scan(tx.QueryRowContext(ctx, stmt))
	}
	_, err = tx.ExecContext(ctx, stmt)
	return err
}

// TestRLSDeniesCrossPrincipalAccess is the heart of task 2.6: one
// installation can never read or write another installation's rows.
func TestRLSDeniesCrossPrincipalAccess(t *testing.T) {
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
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	var visible int
	if err := asTelemetry(ctx, db, "inst-1",
		`SELECT count(*) FROM telemetry_batches`, func(r *sql.Row) error {
			return r.Scan(&visible)
		}); err != nil {
		t.Fatalf("scoped select: %v", err)
	}
	if visible != 1 {
		t.Fatalf("inst-1 must see exactly its batch, saw %d", visible)
	}
	if err := asTelemetry(ctx, db, "inst-1", `INSERT INTO telemetry_batches
		(id, installation_id, sequence) VALUES ('b-x','inst-2',7)`, nil); err == nil {
		t.Fatal("writing another installation's batch must be denied")
	}
	if err := asTelemetry(ctx, db, "inst-1",
		`SELECT count(*) FROM installations`, func(r *sql.Row) error {
			return r.Scan(&visible)
		}); err != nil || visible != 1 {
		t.Fatalf("inst-1 must see only itself in installations: %d (%v)", visible, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE central_telemetry`); err != nil {
		t.Fatalf("role: %v", err)
	}
	var unscoped int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM telemetry_batches`).Scan(&unscoped); err != nil || unscoped != 0 {
		t.Fatalf("without the scope setting every row must be invisible: %d (%v)", unscoped, err)
	}
}
