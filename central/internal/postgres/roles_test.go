package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

// grantCase is one row of the grant/denial matrix of task 2.2.
type grantCase struct {
	role    string
	stmt    string
	allowed bool
}

func grantMatrix() []grantCase {
	return []grantCase{
		// API serves and enrolls, never DDL, never raw telemetry.
		{"central_api", `SELECT count(*) FROM artifacts`, true},
		{"central_api", `INSERT INTO installations (id, token_hash, deletion_hash)
			VALUES ('inst-2', 'th', 'dh')`, true},
		{"central_api", `INSERT INTO audit (actor, action) VALUES ('api', 'enroll')`, true},
		{"central_api", `SELECT count(*) FROM accepted_events`, false},
		{"central_api", `DELETE FROM audit`, false},
		{"central_api", `CREATE TABLE hack (id INT)`, false},
		// Telemetry is append-only inside its scope.
		{"central_telemetry", `INSERT INTO telemetry_batches (id, installation_id, sequence)
			VALUES ('b-2', 'inst-1', 2)`, true},
		{"central_telemetry", `SELECT count(*) FROM installations`, true},
		{"central_telemetry", `UPDATE telemetry_batches SET sequence = 9`, false},
		{"central_telemetry", `SELECT count(*) FROM sources`, false},
		// Worker runs jobs and expires raw events, never drops schema.
		{"central_worker", `UPDATE jobs SET attempts = attempts + 1`, true},
		{"central_worker", `DELETE FROM accepted_events`, true},
		{"central_worker", `DROP TABLE jobs`, false},
		{"central_worker", `DELETE FROM audit`, false},
		// Operations reads evidence, never raw events, never writes.
		{"central_operations", `SELECT count(*) FROM audit`, true},
		{"central_operations", `SELECT count(*) FROM accepted_events`, false},
		{"central_operations", `INSERT INTO audit (actor, action) VALUES ('ops', 'x')`, false},
		// Backup reads everything, writes nothing.
		{"central_backup", `SELECT count(*) FROM accepted_events`, true},
		{"central_backup", `INSERT INTO sources (id, kind) VALUES ('s-2', 'k')`, false},
	}
}

func seed(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		`INSERT INTO sources (id, kind, approved) VALUES ('s-1', 'scryfall', true)`,
		`INSERT INTO installations (id, token_hash, deletion_hash) VALUES ('inst-1','t','d')`,
		`INSERT INTO jobs (id, kind, idempotency_key) VALUES ('j-1','publish','k-1')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func runAs(ctx context.Context, db *sql.DB, role, stmt string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // every probe rolls back
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL ROLE %s", role)); err != nil {
		return err
	}
	if role == "central_telemetry" {
		// The telemetry path always runs installation-scoped (RLS).
		if _, err := tx.ExecContext(ctx,
			`SET LOCAL app.installation_id = 'inst-1'`); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, stmt)
	return err
}
