package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func repoPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := containerDSN(t)
	ctx := context.Background()
	if err := RunMigrator(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := NewPool(ctx, PoolConfig{DSN: dsn, MaxConns: 4,
		AcquireTimeout: 5 * time.Second, QueryTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TestTypedRepositoriesMapStableErrors covers task 2.4's vocabulary:
// duplicate, not-found and cancellation surface as stable errors.
func TestTypedRepositoriesMapStableErrors(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	jobs := JobsRepo{Q: pool}
	job := Job{ID: "j-1", Kind: "publish", IdempotencyKey: "k-1"}
	if err := jobs.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	dup := Job{ID: "j-2", Kind: "publish", IdempotencyKey: "k-1"}
	if err := jobs.Enqueue(ctx, dup); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate idempotency key must map to ErrDuplicate: %v", err)
	}
	if _, err := jobs.Get(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing job must map to ErrNotFound: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := jobs.Get(cancelled, "j-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation must surface untouched: %v", err)
	}
	loaded, err := jobs.Get(ctx, "j-1")
	if err != nil || loaded.Kind != "publish" || loaded.Status != "pending" {
		t.Fatalf("typed roundtrip diverged: %+v (%v)", loaded, err)
	}
}

// TestEnrollWithConsentIsAtomic proves the transactional adapter: the
// duplicate enrollment aborts the whole transaction, so no orphan
// consent receipt survives.
func TestEnrollWithConsentIsAtomic(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	inst := Installation{ID: "inst-1", TokenHash: "t", DeletionHash: "d"}
	if err := EnrollWithConsent(ctx, pool, inst, "product", 1); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if err := EnrollWithConsent(ctx, pool, inst, "product", 2); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("second enrollment must abort as duplicate: %v", err)
	}
	var receipts int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM consent_receipts
		WHERE installation_id = 'inst-1'`).Scan(&receipts); err != nil {
		t.Fatalf("count: %v", err)
	}
	if receipts != 1 { // the aborted transaction left no version-2 receipt
		t.Fatalf("atomicity broken: %d receipts", receipts)
	}
	if _, err := (InstallationsRepo{Q: pool}).ByID(ctx, "inst-1"); err != nil {
		t.Fatalf("by id: %v", err)
	}
}
