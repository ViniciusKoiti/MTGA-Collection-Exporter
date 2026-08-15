package postgres

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func containerDSN(t *testing.T) string {
	t.Helper()
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
			t.Fatalf("CI must run the migrator suite: %v", err)
		}
		t.Skipf("docker unavailable locally; suite is validated in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	return dsn
}

// TestConcurrentMigratorsSerializeUnderTheLock runs two migrators at
// once against a fresh database: the advisory lock serializes them and
// the bookkeeping ends exactly once per migration file.
func TestConcurrentMigratorsSerializeUnderTheLock(t *testing.T) {
	dsn := containerDSN(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = RunMigrator(ctx, dsn)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("migrator %d failed: %v", i, err)
		}
	}
	if err := SchemaReady(ctx, dsn); err != nil {
		t.Fatalf("schema must be ready after migration: %v", err)
	}
}

// TestReadinessRefusesBehindSchemaAndPreflightRefusesAhead covers both
// refusal directions of task 2.5.
func TestReadinessRefusesBehindSchemaAndPreflightRefusesAhead(t *testing.T) {
	dsn := containerDSN(t)
	ctx := context.Background()
	if err := SchemaReady(ctx, dsn); err == nil {
		t.Fatal("readiness must refuse a database with no schema")
	}
	if err := RunMigrator(ctx, dsn); err != nil {
		t.Fatalf("migrator: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations (version)
		VALUES ('9999_from_the_future.sql')`); err != nil {
		t.Fatalf("simulate future version: %v", err)
	}
	err = RunMigrator(ctx, dsn)
	if err == nil || !strings.Contains(err.Error(), "ahead of this binary") {
		t.Fatalf("preflight must refuse a database ahead of the binary: %v", err)
	}
}
