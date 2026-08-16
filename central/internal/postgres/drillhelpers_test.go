package postgres

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// poolFor opens a budgeted pool on an existing drill database.
func poolFor(t *testing.T, dsn string) *pgxpool.Pool {
	t.Helper()
	pool, err := NewPool(context.Background(), PoolConfig{DSN: dsn,
		MaxConns: 4, AcquireTimeout: 5 * time.Second,
		QueryTimeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// drillContainer starts one isolated database for the restore drill;
// skipped locally without Docker, mandatory in CI.
func drillContainer(t *testing.T) (*pgcontainer.PostgresContainer, string) {
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
			t.Fatalf("CI must run the restore drill: %v", err)
		}
		t.Skipf("docker unavailable locally; drill runs in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	return container, dsn
}

// dumpDatabase takes the backup with pg_dump inside the container.
func dumpDatabase(t *testing.T, c *pgcontainer.PostgresContainer) []byte {
	t.Helper()
	code, reader, err := c.Exec(context.Background(),
		[]string{"pg_dump", "-U", "central", "central"}, tcexec.Multiplexed())
	if err != nil || code != 0 {
		t.Fatalf("pg_dump: code %d, %v", code, err)
	}
	dump, err := io.ReadAll(reader)
	if err != nil || len(dump) == 0 {
		t.Fatalf("dump read: %d bytes, %v", len(dump), err)
	}
	return dump
}

// restoreDatabase replays the dump into a fresh container with psql.
func restoreDatabase(t *testing.T, c *pgcontainer.PostgresContainer, sql []byte) {
	t.Helper()
	ctx := context.Background()
	if err := c.CopyToContainer(ctx, sql, "/tmp/restore.sql", 0o644); err != nil {
		t.Fatalf("copy dump: %v", err)
	}
	code, reader, err := c.Exec(ctx, []string{"psql", "-U", "central",
		"-d", "central", "-v", "ON_ERROR_STOP=1", "-f", "/tmp/restore.sql"},
		tcexec.Multiplexed())
	if err != nil || code != 0 {
		out, _ := io.ReadAll(reader)
		t.Fatalf("restore: code %d, %v\n%s", code, err, out)
	}
}
