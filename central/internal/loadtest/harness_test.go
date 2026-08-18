// Package loadtest runs the Stage 1 load suite at CI scale (OpenSpec
// add-central-go-platform, task 8.1) against the REAL HTTP API wired
// to a REAL PostgreSQL container. The 10k-DAU assumptions and the
// cost extrapolation live in docs/load-stage1.md.
package loadtest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/httpapi"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// startStack boots the Stage 1 stack with its 8-connection budget.
func startStack(t *testing.T) (string, *pgxpool.Pool) {
	return startStackWith(t, 8)
}

// startStackWith boots container, schema, a pool with the given
// connection budget, and the real routers.
func startStackWith(t *testing.T, maxConns int) (string, *pgxpool.Pool) {
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
			t.Fatalf("CI must run the load suite: %v", err)
		}
		t.Skipf("docker unavailable locally; load runs in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	if err := postgres.RunMigrator(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := postgres.NewPool(ctx, postgres.PoolConfig{DSN: dsn,
		MaxConns: maxConns, AcquireTimeout: 5 * time.Second,
		QueryTimeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	repo := postgres.InstallationsRepo{Q: pool}
	if err := repo.Enroll(ctx, postgres.Installation{ID: "inst-load",
		TokenHash: "h-load", DeletionHash: "d-load"}); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	verify := func(context.Context, string) (httpapi.Principal, error) {
		return httpapi.Principal{InstallationID: "inst-load",
			Scope: "installation"}, nil
	}
	manifest := &httpapi.DocHandler{Source: func(context.Context) (httpapi.Document, error) {
		return httpapi.Document{Body: []byte(`{"schema":"cards-v1"}`),
			ContentType: "application/json"}, nil
	}, TTL: time.Minute, MaxAge: 60}
	mux := http.NewServeMux()
	mux.Handle("/v1/manifest/current", manifest)
	mux.Handle("/v1/telemetry", httpapi.TelemetryRoutes(
		httpapi.EventPolicy{Names: map[string]bool{"scan_completed": true},
			MaxEvents: 4, MaxAttrs: 2, MaxValueLen: 32},
		pgTelemetry{pool: pool}, verify,
		&httpapi.RatePolicy{Limit: 1_000_000, Window: time.Minute, MaxKeys: 8},
		func(httpapi.Observation) {}))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server.URL, pool
}
