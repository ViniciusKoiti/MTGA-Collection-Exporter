package publication

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/concurrency"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/platform/config"
)

// publish é a fase estagiada e sequencial da publicação: lotes limitados no
// banco, snapshot imutável no storage, assinatura e só então a ativação
// atômica. Qualquer falha antes de Activate deixa o manifesto anterior como
// corrente (requisito "Publication is atomic" da spec catalog-publication).
func publish(
	ctx context.Context,
	b config.Budgets,
	schema string,
	cards []catalog.Card,
	deps Deps,
) (catalog.Manifest, error) {
	for _, batch := range concurrency.Batches(cards, b.BatchSize) {
		if err := deps.Writer.WriteBatch(ctx, batch); err != nil {
			return catalog.Manifest{}, err
		}
	}
	snapshot := catalog.Snapshot{SchemaVersion: schema, Cards: cards}
	ref, err := deps.Objects.WriteSnapshot(ctx, snapshot)
	if err != nil {
		return catalog.Manifest{}, err
	}
	manifest, err := deps.Signer.Sign(ctx, ref)
	if err != nil {
		return catalog.Manifest{}, err
	}
	if err := deps.Activator.Activate(ctx, manifest); err != nil {
		return catalog.Manifest{}, err
	}
	return manifest, nil
}
