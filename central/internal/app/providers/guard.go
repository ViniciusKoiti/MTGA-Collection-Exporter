package providers

import (
	"context"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/ports"
)

// Guard wraps a fetcher so EVERY job passes the registry before any
// upstream call: a provider disabled mid-run stops on the next job,
// which is exactly the rapid-disablement contract.
type Guard struct {
	Registry *Registry
	Next     ports.ProviderFetcher
	Clock    func() time.Time
}

// Fetch authorizes the job and only then delegates to the adapter.
func (g Guard) Fetch(ctx context.Context,
	job catalog.ProviderJob) ([]catalog.Observation, error) {
	if _, err := g.Registry.Authorize(job.Provider, job.Kind, g.Clock()); err != nil {
		return nil, err
	}
	return g.Next.Fetch(ctx, job)
}

// compile-time proof that Guard satisfies the adapter contract.
var _ ports.ProviderFetcher = Guard{}
