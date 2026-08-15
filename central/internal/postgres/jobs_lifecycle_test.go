package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestJobLifecycleWithRetryPolicy walks the full state machine: claim,
// heartbeat, retryable failure back to pending, and both terminal
// outcomes under the attempts budget.
func TestJobLifecycleWithRetryPolicy(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	jobs := JobsRepo{Q: pool}
	if err := jobs.Enqueue(ctx, Job{ID: "j-1", Kind: "publish",
		IdempotencyKey: "k-1"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	first, err := jobs.ClaimNext(ctx, "w1", time.Minute)
	if err != nil || first.ID != "j-1" || first.Attempts != 1 {
		t.Fatalf("first claim: %+v (%v)", first, err)
	}
	if err := jobs.Heartbeat(ctx, "j-1", "w1", time.Minute); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	terminal, err := jobs.Fail(ctx, "j-1", "w1", 3)
	if err != nil || terminal {
		t.Fatalf("first failure must be retryable: terminal=%t (%v)", terminal, err)
	}

	second, err := jobs.ClaimNext(ctx, "w2", time.Minute)
	if err != nil || second.Attempts != 2 {
		t.Fatalf("redelivery: %+v (%v)", second, err)
	}
	if _, err := jobs.Fail(ctx, "j-1", "w2", 3); err != nil {
		t.Fatalf("second failure: %v", err)
	}
	third, err := jobs.ClaimNext(ctx, "w1", time.Minute)
	if err != nil || third.Attempts != 3 {
		t.Fatalf("third claim: %+v (%v)", third, err)
	}
	terminal, err = jobs.Fail(ctx, "j-1", "w1", 3)
	if err != nil || !terminal {
		t.Fatalf("attempts budget reached must be terminal: %t (%v)", terminal, err)
	}
	if _, err := jobs.ClaimNext(ctx, "w1", time.Minute); !errors.Is(err, ErrNotFound) {
		t.Fatalf("terminal jobs are never redelivered: %v", err)
	}

	if err := jobs.Enqueue(ctx, Job{ID: "j-2", Kind: "publish",
		IdempotencyKey: "k-2"}); err != nil {
		t.Fatalf("enqueue 2: %v", err)
	}
	if _, err := jobs.ClaimNext(ctx, "w1", time.Minute); err != nil {
		t.Fatalf("claim 2: %v", err)
	}
	if err := jobs.Complete(ctx, "j-2", "w1"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	done, err := jobs.Get(ctx, "j-2")
	if err != nil || done.Status != JobSucceeded {
		t.Fatalf("success must be terminal: %+v (%v)", done, err)
	}
}

// TestDeadWorkerLeaseExpiryAndReclaim: a worker that stops heartbeating
// loses the job; another worker reclaims it and the dead owner can no
// longer heartbeat, complete, or produce effects.
func TestDeadWorkerLeaseExpiryAndReclaim(t *testing.T) {
	pool := repoPool(t)
	ctx := context.Background()
	jobs := JobsRepo{Q: pool}
	if err := jobs.Enqueue(ctx, Job{ID: "j-1", Kind: "publish",
		IdempotencyKey: "k-1"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if _, err := jobs.ClaimNext(ctx, "dead", 100*time.Millisecond); err != nil {
		t.Fatalf("claim: %v", err)
	}
	time.Sleep(300 * time.Millisecond) // the dead worker never heartbeats

	reclaimed, err := jobs.ClaimNext(ctx, "alive", time.Minute)
	if err != nil || reclaimed.Attempts != 2 {
		t.Fatalf("expired lease must be reclaimable: %+v (%v)", reclaimed, err)
	}
	if err := jobs.Heartbeat(ctx, "j-1", "dead", time.Minute); !errors.Is(err, ErrNotFound) {
		t.Fatalf("dead owner heartbeat must report the lost lease: %v", err)
	}
	if err := jobs.Complete(ctx, "j-1", "dead"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("dead owner completion must be refused: %v", err)
	}
	if err := jobs.Complete(ctx, "j-1", "alive"); err != nil {
		t.Fatalf("the live owner finishes normally: %v", err)
	}
}
