// Package publication orchestrates the bounded catalog-publication
// pipeline (docs/architecture/mtga-go-concurrency.puml):
//
//	provider jobs -> fetch pool -> observations -> normalize pool ->
//	validate/quarantine filter -> stable reducer -> bounded batches +
//	canonical snapshot -> object writer -> signer -> atomic activation
//
// The run owner (Run) creates the errgroup, is the only one waiting on
// every goroutine, and propagates the first error via ctx cancellation.
package publication

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/concurrency"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/ports"
)

// Deps groups the adapters publication requires.
type Deps struct {
	Fetcher    ports.ProviderFetcher
	Normalize  ports.Normalizer
	Validator  ports.CardValidator
	Quarantine ports.QuarantineSink
	Writer     ports.BatchWriter
	Objects    ports.ObjectWriter
	Signer     ports.Signer
	Activator  ports.Activator
}

// Run executes one full publication and returns the activated
// manifest. The parallel phase ends at the reducer; batch writes,
// snapshot, signing and activation stay sequential so the publication
// remains atomic and auditable.
func Run(
	ctx context.Context,
	budgets config.Budgets,
	schema string,
	jobs []catalog.ProviderJob,
	deps Deps,
) (catalog.Manifest, error) {
	cards, err := gather(ctx, budgets, jobs, deps)
	if err != nil {
		return catalog.Manifest{}, err
	}
	return publish(ctx, budgets, schema, cards, deps)
}

// gather is the bounded parallel phase: fetch -> normalize ->
// validate/quarantine -> stable reduction. Unsound cards leave the
// pipeline through the quarantine sink instead of failing the run.
func gather(
	ctx context.Context,
	b config.Budgets,
	jobs []catalog.ProviderJob,
	deps Deps,
) ([]catalog.Card, error) {
	g, gctx := errgroup.WithContext(ctx)
	jobsCh := concurrency.Source(gctx, g, jobs, b.ProviderBuffer)
	observations := concurrency.FlatPool(gctx, g, jobsCh, b.FetchWorkers,
		b.ObservationBuffer, deps.Fetcher.Fetch)
	normalized := concurrency.Pool(gctx, g, observations, b.NormalizeWorkers,
		b.NormalizedBuffer, deps.Normalize.Normalize)
	validated := concurrency.Filter(gctx, g, normalized, b.NormalizedBuffer,
		func(ctx context.Context, card catalog.Card) (bool, error) {
			if err := deps.Validator.Check(ctx, card); err != nil {
				return false, deps.Quarantine.Quarantine(ctx, card, err.Error())
			}
			return true, nil
		})

	var cards []catalog.Card
	g.Go(func() error {
		var err error
		cards, err = concurrency.Reduce(gctx, validated,
			func(a, b catalog.Card) bool { return a.GrpID < b.GrpID })
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return cards, nil
}
