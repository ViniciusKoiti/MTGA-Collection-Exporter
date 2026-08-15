package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// expectedTables is the task 2.1 checklist: every domain the central
// schema must own from day one.
var expectedTables = []string{
	"sources", "snapshots", "artifacts", "installations", "consent_receipts",
	"telemetry_batches", "accepted_events", "aggregates", "deletion_requests",
	"jobs", "outbox", "audit", "schema_migrations",
}

// TestForwardMigrationsApplyAndAreIdempotent runs against a real
// PostgreSQL container: skipped locally without Docker, mandatory in CI.
func TestForwardMigrationsApplyAndAreIdempotent(t *testing.T) {
	ctx := context.Background()
	container, err := pgcontainer.Run(ctx, "postgres:16-alpine",
		pgcontainer.WithDatabase("central"),
		pgcontainer.WithUsername("central"),
		pgcontainer.WithPassword("central"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(90*time.Second)))
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("CI must run the migration suite: %v", err)
		}
		t.Skipf("docker unavailable locally; suite is validated in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}

	if err := Migrate(ctx, dsn); err != nil {
		t.Fatalf("first migration pass: %v", err)
	}
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatalf("second pass must be idempotent: %v", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, table := range expectedTables {
		var found int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1`, table).Scan(&found); err != nil {
			t.Fatalf("check %s: %v", table, err)
		}
		if found != 1 {
			t.Fatalf("table %s missing after migrations", table)
		}
	}
	var applied int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("versions: %v", err)
	}
	files, err := migrations.ReadDir("migrations")
	if err != nil {
		t.Fatalf("embedded migrations: %v", err)
	}
	if applied != len(files) { // idempotency: every file applied exactly once
		t.Fatalf("expected %d applied versions, got %d", len(files), applied)
	}
}
