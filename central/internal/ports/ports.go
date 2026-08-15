// Package ports declara os contratos que o núcleo da aplicação exige dos
// adaptadores (provedores, PostgreSQL, object storage, assinatura). Os
// adaptadores reais ficam em internal/adapters; testes usam fakes.
package ports

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// ProviderFetcher busca observações de um provedor aprovado, respeitando o
// orçamento de rate/timeout do adaptador. Deve honrar cancelamento via ctx.
type ProviderFetcher interface {
	Fetch(ctx context.Context, job catalog.ProviderJob) ([]catalog.Observation, error)
}

// Normalizer converte uma observação bruta em carta validada.
type Normalizer interface {
	Normalize(ctx context.Context, obs catalog.Observation) (catalog.Card, error)
}

// BatchWriter persiste cartas em lotes limitados dentro do orçamento do pgxpool.
type BatchWriter interface {
	WriteBatch(ctx context.Context, cards []catalog.Card) error
}

// ObjectWriter grava o snapshot canônico como objeto imutável e devolve a
// referência com hash verificado pós-upload.
type ObjectWriter interface {
	WriteSnapshot(ctx context.Context, snap catalog.Snapshot) (catalog.ObjectRef, error)
}

// Signer assina o manifesto do objeto publicado (Ed25519 com key ID).
type Signer interface {
	Sign(ctx context.Context, ref catalog.ObjectRef) (catalog.Manifest, error)
}

// Activator torna o manifesto assinado o snapshot corrente em uma transação;
// em caso de falha o manifesto anterior permanece o corrente.
type Activator interface {
	Activate(ctx context.Context, man catalog.Manifest) error
}
