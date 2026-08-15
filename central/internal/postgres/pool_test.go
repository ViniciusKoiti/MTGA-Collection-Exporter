package postgres

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestPoolEnforcesBudgetsAndTimeouts proves the three budget walls on a
// real PostgreSQL: connection ceiling (saturated acquire times out and
// is visible in health), server-side statement timeout, and rejection of
// incomplete budgets. Skipped locally without Docker, mandatory in CI.
func TestPoolEnforcesBudgetsAndTimeouts(t *testing.T) {
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
			t.Fatalf("CI must run the pool suite: %v", err)
		}
		t.Skipf("docker unavailable locally; suite is validated in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}

	if _, err := NewPool(ctx, PoolConfig{DSN: dsn, MaxConns: 0,
		AcquireTimeout: time.Second, QueryTimeout: time.Second}); err == nil {
		t.Fatal("incomplete budget must be rejected")
	}

	pool, err := NewPool(ctx, PoolConfig{DSN: dsn, MaxConns: 2,
		AcquireTimeout: 2 * time.Second, QueryTimeout: 500 * time.Millisecond})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	first, err := Acquire(ctx, pool, 2*time.Second)
	if err != nil {
		t.Fatalf("acquire 1: %v", err)
	}
	defer first.Release()
	second, err := Acquire(ctx, pool, 2*time.Second)
	if err != nil {
		t.Fatalf("acquire 2: %v", err)
	}
	defer second.Release()

	start := time.Now()
	if _, err := Acquire(ctx, pool, time.Second); err == nil {
		t.Fatal("third acquire must hit the ceiling")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("saturated acquire must respect its timeout")
	}
	if Health(pool).Total > 2 {
		t.Fatalf("ceiling breached: %+v", Health(pool))
	}

	_, err = first.Exec(ctx, "SELECT pg_sleep(5)")
	if err == nil || !strings.Contains(err.Error(), "statement timeout") {
		t.Fatalf("server-side statement timeout expected, got: %v", err)
	}
}
