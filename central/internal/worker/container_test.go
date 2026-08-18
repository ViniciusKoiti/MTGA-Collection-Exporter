package worker

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// repoPoolForWorker spins the container database for the poller suite;
// skipped locally without Docker, mandatory in CI.
func repoPoolForWorker(t *testing.T) *pgxpool.Pool {
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
			t.Fatalf("CI must run the poller suite: %v", err)
		}
		t.Skipf("docker unavailable locally; suite is validated in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	if err := postgres.RunMigrator(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := postgres.NewPool(ctx, postgres.PoolConfig{DSN: dsn, MaxConns: 4,
		AcquireTimeout: 5 * time.Second, QueryTimeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
