package publication

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// instrumentedDeps wraps fakeDeps to measure concurrency and lag.
type instrumentedDeps struct {
	*fakeDeps
	inFetch  atomic.Int64
	maxFetch atomic.Int64
	produced atomic.Int64
	consumed atomic.Int64
	maxLag   atomic.Int64
	slow     time.Duration
}

func storeMax(target *atomic.Int64, value int64) {
	for {
		current := target.Load()
		if value <= current || target.CompareAndSwap(current, value) {
			return
		}
	}
}

func (d *instrumentedDeps) Fetch(ctx context.Context,
	job catalog.ProviderJob) ([]catalog.Observation, error) {
	storeMax(&d.maxFetch, d.inFetch.Add(1))
	defer d.inFetch.Add(-1)
	obs, err := d.fakeDeps.Fetch(ctx, job)
	d.produced.Add(int64(len(obs)))
	return obs, err
}

func (d *instrumentedDeps) Normalize(ctx context.Context,
	obs catalog.Observation) (catalog.Card, error) {
	storeMax(&d.maxLag, d.produced.Load()-d.consumed.Load())
	if d.slow > 0 {
		time.Sleep(d.slow)
	}
	d.consumed.Add(1)
	return d.fakeDeps.Normalize(ctx, obs)
}

func (d *instrumentedDeps) deps() Deps {
	base := d.fakeDeps.deps()
	base.Fetcher, base.Normalize = d, d
	return base
}

// TestFetchConcurrencyStaysWithinTheBudget: the pool never runs more
// simultaneous fetches than FetchWorkers allows.
func TestFetchConcurrencyStaysWithinTheBudget(t *testing.T) {
	inst := &instrumentedDeps{fakeDeps: newFakeDeps(25, nil),
		slow: time.Millisecond}
	if _, err := Run(t.Context(), budgets(), "v1", jobs(8), inst.deps()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if max := inst.maxFetch.Load(); max > int64(budgets().FetchWorkers) {
		t.Fatalf("fetch concurrency %d exceeded the budget %d",
			max, budgets().FetchWorkers)
	}
}

// TestBackpressureBoundsInFlightObservations: with a slow consumer the
// producers block instead of running ahead — in-flight observations
// stay within workers*batch + buffer, far below the total.
func TestBackpressureBoundsInFlightObservations(t *testing.T) {
	inst := &instrumentedDeps{fakeDeps: newFakeDeps(25, nil),
		slow: time.Millisecond}
	if _, err := Run(t.Context(), budgets(), "v1", jobs(8), inst.deps()); err != nil {
		t.Fatalf("run: %v", err)
	}
	b := budgets()
	bound := int64(b.FetchWorkers*25 + b.ObservationBuffer + b.NormalizedBuffer)
	if lag := inst.maxLag.Load(); lag > bound {
		t.Fatalf("backpressure failed: %d observations in flight, bound %d",
			lag, bound)
	}
}
