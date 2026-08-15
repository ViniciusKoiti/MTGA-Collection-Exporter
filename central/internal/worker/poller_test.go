package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// pollerFor builds a poller over the container database with a handler
// that records processed IDs and fails the IDs listed in failing.
func pollerFor(t *testing.T, jobs postgres.JobsRepo, workers int,
	processed *sync.Map, failing map[string]bool) Poller {
	t.Helper()
	return Poller{
		Jobs:  jobs,
		Owner: "w-test",
		Handle: func(_ context.Context, job postgres.Job) error {
			processed.Store(job.ID, true)
			if failing[job.ID] {
				return errors.New("handler failure")
			}
			return nil
		},
		Workers: workers, DBBudget: 4,
		Lease: 5 * time.Second, Heartbeat: time.Second,
		PollEvery: 100 * time.Millisecond, MaxAttempts: 2,
	}
}

// TestPollerProcessesQueueWithinBudgets drains a queue with a bounded
// worker group, applies the retry policy and stops cleanly on cancel.
func TestPollerProcessesQueueWithinBudgets(t *testing.T) {
	pool := repoPoolForWorker(t)
	jobs := postgres.JobsRepo{Q: pool}
	ctx := context.Background()
	for i := range 5 {
		if err := jobs.Enqueue(ctx, postgres.Job{ID: fmt.Sprintf("j-%d", i),
			Kind: "publish", IdempotencyKey: fmt.Sprintf("k-%d", i)}); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}
	var processed sync.Map
	poller := pollerFor(t, jobs, 2, &processed, map[string]bool{"j-3": true})

	runCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- poller.Run(runCtx) }()

	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		if terminalCount(t, jobs, ctx) == 5 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("poller must stop cleanly on cancel: %v", err)
	}
	for i := range 5 {
		id := fmt.Sprintf("j-%d", i)
		job, err := jobs.Get(ctx, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		want := postgres.JobSucceeded
		if id == "j-3" {
			want = postgres.JobFailed // retried to the attempts budget
			if job.Attempts != 2 {
				t.Fatalf("j-3 must exhaust its 2 attempts: %+v", job)
			}
		}
		if job.Status != want {
			t.Fatalf("%s: expected %s, got %+v", id, want, job)
		}
	}
}
