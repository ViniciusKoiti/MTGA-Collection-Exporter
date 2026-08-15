package publication

import (
	"runtime"
	"testing"
	"time"
)

// TestRunLeaksNoGoroutines: after Run returns, every pipeline
// goroutine is gone — Run owns its whole errgroup.
func TestRunLeaksNoGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 3 {
		fake := newFakeDeps(10, nil)
		if _, err := Run(t.Context(), budgets(), "v1", jobs(4), fake.deps()); err != nil {
			t.Fatalf("run: %v", err)
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+2 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("goroutines leaked: before %d, after %d",
		before, runtime.NumGoroutine())
}

// TestConcurrentRunsStayIndependent: each Run owns its own errgroup
// and state — two publications at once never cross results.
func TestConcurrentRunsStayIndependent(t *testing.T) {
	first := newFakeDeps(10, nil)
	second := newFakeDeps(10, nil)
	errs := make(chan error, 2)
	counts := make(chan int, 2)
	for _, fake := range []*fakeDeps{first, second} {
		go func() {
			_, err := Run(t.Context(), budgets(), "v1", jobs(3), fake.deps())
			errs <- err
			counts <- len(fake.cartasEscritas())
		}()
	}
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent run: %v", err)
		}
		if got := <-counts; got != 30 {
			t.Fatalf("each run must own exactly its 30 cards, got %d", got)
		}
	}
	if first.ativacoes != 1 || second.ativacoes != 1 {
		t.Fatalf("each run must activate exactly once: %d/%d",
			first.ativacoes, second.ativacoes)
	}
}
