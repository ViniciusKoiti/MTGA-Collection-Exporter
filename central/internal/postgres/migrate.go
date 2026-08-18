// Package postgres owns the central database migrations (OpenSpec
// add-central-go-platform, task 2.1): forward-only SQL files embedded in
// the binary, applied one per transaction and recorded in
// schema_migrations. The dedicated migrator command and role arrive with
// task 2.5; this runner is what it will call.
package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies every pending forward migration in filename order.
// Re-running is idempotent: applied versions are skipped.
func Migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("postgres: open: %w", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("postgres: bootstrap: %w", err)
	}
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names) // deterministic forward order
	for _, name := range names {
		if err := apply(ctx, db, name); err != nil {
			return err
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, name string) error {
	var seen int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, name).Scan(&seen); err != nil {
		return err
	}
	if seen > 0 {
		return nil
	}
	payload, err := migrations.ReadFile("migrations/" + name)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, string(payload)); err != nil {
		return fmt.Errorf("postgres: migration %s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
		return err
	}
	return tx.Commit()
}
