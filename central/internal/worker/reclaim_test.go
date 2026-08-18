package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// TestWorkerDeathLeaseExpiryAndDuplicateEffectFencing: a dead worker's
// lease protects the job until it expires, the reclaim bumps attempts,
// and the dead owner is fenced out of every later effect.
func TestWorkerDeathLeaseExpiryAndDuplicateEffectFencing(t *testing.T) {
	pool := repoPoolForWorker(t)
	jobs := postgres.JobsRepo{Q: pool}
	ctx := context.Background()
	if err := jobs.Enqueue(ctx, postgres.Job{ID: "dead-1", Kind: "publish",
		IdempotencyKey: "dead-1"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	first, err := jobs.ClaimNext(ctx, "owner-a", 300*time.Millisecond)
	if err != nil || first.ID != "dead-1" || first.Attempts != 1 {
		t.Fatalf("first claim: %+v %v", first, err)
	}
	// owner-a dies silently: no heartbeat. While the lease lives, the
	// job must stay protected from other owners.
	if _, err := jobs.ClaimNext(ctx, "owner-b", time.Second); !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("a live lease must protect the job: %v", err)
	}
	time.Sleep(400 * time.Millisecond) // the lease expires
	second, err := jobs.ClaimNext(ctx, "owner-b", time.Second)
	if err != nil || second.ID != "dead-1" || second.Attempts != 2 {
		t.Fatalf("expired lease must be reclaimed with attempts bumped: %+v %v",
			second, err)
	}
	// The dead owner comes back: every effect it tries is fenced.
	if err := jobs.Heartbeat(ctx, "dead-1", "owner-a", time.Second); !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("stale heartbeat must be fenced: %v", err)
	}
	if err := jobs.Complete(ctx, "dead-1", "owner-a"); !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("stale completion must be fenced: %v", err)
	}
	if err := jobs.Complete(ctx, "dead-1", "owner-b"); err != nil {
		t.Fatalf("the live owner must complete: %v", err)
	}
	if err := jobs.Complete(ctx, "dead-1", "owner-b"); !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("completing twice must be impossible: %v", err)
	}
	job, err := jobs.Get(ctx, "dead-1")
	if err != nil || job.Status != postgres.JobSucceeded || job.Attempts != 2 {
		t.Fatalf("exactly one effect must survive: %+v %v", job, err)
	}
}
