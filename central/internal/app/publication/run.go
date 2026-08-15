// Package publication orquestra o pipeline limitado de publicação de catálogo
// (docs/architecture/mtga-go-concurrency.puml):
//
//	provider jobs -> fetch pool -> observations -> normalize pool ->
//	normalized -> redutor estável -> lotes limitados + snapshot canônico ->
//	object writer -> signer -> ativação atômica
//
// O dono da execução (Run) cria o errgroup, é o único a esperar todas as
// goroutines e propaga o primeiro erro via cancelamento do contexto.
package publication

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/concurrency"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/ports"
)

// Deps agrupa os adaptadores exigidos pela publicação.
type Deps struct {
	Fetcher   ports.ProviderFetcher
	Normalize ports.Normalizer
	Writer    ports.BatchWriter
	Objects   ports.ObjectWriter
	Signer    ports.Signer
	Activator ports.Activator
}

// Run executa uma publicação completa e devolve o manifesto ativado.
// A fase paralela termina no redutor; escrita, snapshot, assinatura e
// ativação são sequenciais para manter a publicação atômica e auditável.
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

// gather é a fase paralela limitada: fetch -> normalize -> redução estável.
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

	var cards []catalog.Card
	g.Go(func() error {
		var err error
		cards, err = concurrency.Reduce(gctx, normalized,
			func(a, b catalog.Card) bool { return a.GrpID < b.GrpID })
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return cards, nil
}
