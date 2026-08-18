// Package worker runs the durable-job poller (OpenSpec
// add-central-go-platform, task 6.2): ONE context-owned polling loop
// feeds a bounded worker group whose size can never exceed the database
// budget. All goroutines and channels come from the concurrency package;
// the first error cancels the shared context and Run waits for every
// owned goroutine.
package worker

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/concurrency"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

// Handler processes one claimed job; returning an error triggers the
// retry policy.
type Handler func(ctx context.Context, job postgres.Job) error

// Poller owns claim, dispatch, heartbeat and outcome recording.
type Poller struct {
	Jobs        postgres.JobsRepo
	Handle      Handler
	Owner       string
	Workers     int
	DBBudget    int // pool ceiling of this process; workers may not exceed it
	Lease       time.Duration
	Heartbeat   time.Duration
	PollEvery   time.Duration
	MaxAttempts int
}

func (p Poller) validate() error {
	if p.Workers < 1 || p.Workers > p.DBBudget {
		return fmt.Errorf("worker: %d workers exceed the database budget %d",
			p.Workers, p.DBBudget)
	}
	if p.Lease <= 0 || p.Heartbeat <= 0 || p.PollEvery <= 0 || p.MaxAttempts < 1 {
		return fmt.Errorf("worker: poller budgets incomplete: %+v", p)
	}
	return nil
}

// Run polls and processes jobs until the context is cancelled; it
// returns only after every owned goroutine finished.
func (p Poller) Run(ctx context.Context) error {
	if err := p.validate(); err != nil {
		return err
	}
	g, gctx := errgroup.WithContext(ctx)
	jobs := concurrency.Poll(gctx, g, p.PollEvery, p.Workers,
		func(ctx context.Context) (postgres.Job, bool, error) {
			job, err := p.Jobs.ClaimNext(ctx, p.Owner, p.Lease)
			if err != nil {
				if isNotFound(err) {
					return postgres.Job{}, false, nil // queue drained
				}
				return postgres.Job{}, false, err
			}
			return job, true, nil
		})
	results := concurrency.Pool(gctx, g, jobs, p.Workers, p.Workers,
		func(ctx context.Context, job postgres.Job) (string, error) {
			return job.ID, p.process(ctx, job)
		})
	g.Go(func() error {
		_, err := concurrency.Reduce(gctx, results,
			func(a, b string) bool { return a < b })
		return err
	})
	return g.Wait()
}

// process runs the handler under the heartbeat and records the outcome.
func (p Poller) process(ctx context.Context, job postgres.Job) error {
	workErr := concurrency.WithHeartbeat(ctx, p.Heartbeat,
		func(ctx context.Context) error {
			return p.Jobs.Heartbeat(ctx, job.ID, p.Owner, p.Lease)
		},
		func(ctx context.Context) error {
			return p.Handle(ctx, job)
		})
	if workErr == nil {
		return p.Jobs.Complete(ctx, job.ID, p.Owner)
	}
	if _, err := p.Jobs.Fail(ctx, job.ID, p.Owner, p.MaxAttempts); err != nil {
		return err // lost lease: the reclaim already owns the job
	}
	return nil // handled by the retry policy; the poller keeps running
}
