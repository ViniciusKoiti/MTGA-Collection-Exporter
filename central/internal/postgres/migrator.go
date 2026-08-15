package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// migrationLockKey serializes migrators cluster-wide via advisory lock.
const migrationLockKey = int64(0x6d746761) // "mtga"

// RunMigrator is the dedicated migration entry point (task 2.5): it
// takes the cluster-wide advisory lock, refuses to run against a
// database AHEAD of this binary (compatibility preflight), applies the
// forward migrations and verifies the result. API replicas never call
// this — they only ever call SchemaReady.
func RunMigrator(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("postgres: migrator open: %w", err)
	}
	defer func() { _ = db.Close() }()
	lock, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	// Blocks until any concurrent migrator finishes; the lock lives on
	// this dedicated session for the whole run.
	if _, err := lock.ExecContext(ctx,
		`SELECT pg_advisory_lock($1)`, migrationLockKey); err != nil {
		return fmt.Errorf("postgres: migration lock: %w", err)
	}
	defer func() {
		_, _ = lock.ExecContext(context.WithoutCancel(ctx),
			`SELECT pg_advisory_unlock($1)`, migrationLockKey)
	}()
	if err := preflight(ctx, db); err != nil {
		return err
	}
	if err := Migrate(ctx, dsn); err != nil {
		return err
	}
	return SchemaReady(ctx, dsn)
}

// preflight refuses a database that knows migrations this binary does
// not ship: that means the binary is older than the schema and must not
// touch it.
func preflight(ctx context.Context, db *sql.DB) error {
	known := map[string]bool{}
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		known[entry.Name()] = true
	}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil // no bookkeeping table yet: fresh database
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		if !known[version] {
			return fmt.Errorf(
				"postgres: database is ahead of this binary (unknown migration %s)", version)
		}
	}
	return rows.Err()
}

// SchemaReady is the API readiness gate: it fails while any embedded
// migration has not been applied, so replicas refuse traffic on an
// outdated schema instead of serving wrong answers.
func SchemaReady(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	var applied int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		return fmt.Errorf("postgres: schema not ready: %w", err)
	}
	if applied < len(entries) {
		return fmt.Errorf("postgres: schema behind: %d of %d migrations applied",
			applied, len(entries))
	}
	return nil
}
