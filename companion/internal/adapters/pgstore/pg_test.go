package pgstore

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/storetest"
)

// TestRunStorePostgresContractSuite runs the SAME contract suite the
// in-memory and SQLite adapters pass, against a real PostgreSQL spun up
// by Testcontainers. Without Docker locally the test SKIPS; in CI
// (ubuntu runners ship Docker) it is mandatory and fails loudly.
func TestRunStorePostgresContractSuite(t *testing.T) {
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("wf"),
		postgres.WithUsername("wf"),
		postgres.WithPassword("wf"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(90*time.Second)))
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("CI must run the PostgreSQL suite: %v", err)
		}
		t.Skipf("docker unavailable locally; suite is validated in CI: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	store, err := Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	storetest.Executa(t, func(t *testing.T) wf.RunStore {
		if err := store.reset(context.Background()); err != nil {
			t.Fatalf("reset: %v", err)
		}
		return store
	})
}
