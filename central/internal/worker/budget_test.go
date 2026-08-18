package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/postgres"
)

func terminalCount(t *testing.T, jobs postgres.JobsRepo, ctx context.Context) int {
	t.Helper()
	count := 0
	for i := range 5 {
		job, err := jobs.Get(ctx, fmt.Sprintf("j-%d", i))
		if err == nil && (job.Status == postgres.JobSucceeded ||
			job.Status == postgres.JobFailed) {
			count++
		}
	}
	return count
}

// TestPollerRefusesWorkersBeyondTheDatabaseBudget: the worker group can
// never exceed what the database pool reserved for this process.
func TestPollerRefusesWorkersBeyondTheDatabaseBudget(t *testing.T) {
	poller := Poller{Workers: 8, DBBudget: 4, Lease: time.Second,
		Heartbeat: time.Second, PollEvery: time.Second, MaxAttempts: 1}
	if err := poller.Run(context.Background()); err == nil {
		t.Fatal("workers beyond the database budget must be refused")
	}
	zeroed := Poller{Workers: 1, DBBudget: 4}
	if err := zeroed.Run(context.Background()); err == nil {
		t.Fatal("incomplete poller budgets must be refused")
	}
}
