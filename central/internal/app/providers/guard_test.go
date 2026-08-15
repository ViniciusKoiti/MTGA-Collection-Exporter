package providers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

type fetchCounter struct{ calls int }

func (f *fetchCounter) Fetch(_ context.Context,
	job catalog.ProviderJob) ([]catalog.Observation, error) {
	f.calls++
	return []catalog.Observation{{Provider: job.Provider, GrpID: 1}}, nil
}

// TestGuardStopsDisabledProvidersMidRun: the upstream adapter is never
// reached once the provider is cut off between jobs.
func TestGuardStopsDisabledProvidersMidRun(t *testing.T) {
	reg := scryfallOnly(t, 10)
	upstream := &fetchCounter{}
	guard := Guard{Registry: reg, Next: upstream,
		Clock: func() time.Time { return time.Unix(1000, 0) }}
	job := catalog.ProviderJob{Provider: "scryfall", Kind: "cards"}

	if _, err := guard.Fetch(context.Background(), job); err != nil {
		t.Fatalf("approved job must fetch: %v", err)
	}
	reg.Disable("scryfall")
	if _, err := guard.Fetch(context.Background(), job); !errors.Is(err, ErrDisabled) {
		t.Fatalf("disabled provider must stop before the adapter: %v", err)
	}
	if upstream.calls != 1 {
		t.Fatalf("the adapter must not see refused jobs: %d calls", upstream.calls)
	}
}

// TestGuardEnforcesTheRateBudgetBeforeTheAdapter: the rate refusal
// happens in the guard, not the upstream.
func TestGuardEnforcesTheRateBudgetBeforeTheAdapter(t *testing.T) {
	reg := scryfallOnly(t, 1)
	upstream := &fetchCounter{}
	guard := Guard{Registry: reg, Next: upstream,
		Clock: func() time.Time { return time.Unix(1000, 0) }}
	job := catalog.ProviderJob{Provider: "scryfall", Kind: "cards"}

	if _, err := guard.Fetch(context.Background(), job); err != nil {
		t.Fatalf("first job within budget: %v", err)
	}
	if _, err := guard.Fetch(context.Background(), job); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("budget spent must refuse: %v", err)
	}
	if upstream.calls != 1 {
		t.Fatalf("refused jobs must not reach the adapter: %d", upstream.calls)
	}
}
