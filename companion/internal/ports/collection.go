// Package ports declara os contratos que a aplicação do companion exige
// dos adaptadores (tarefa 2.3 do OpenSpec introduce-agentic-go-companion).
// Adaptadores reais vivem em internal/adapters; cada um mantém asserção de
// implementação em tempo de compilação.
package ports

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// CollectionSource observa a coleção em uma fonte confiável (import JSON,
// Detailed Logs aceitos ou a ponte legada explícita).
type CollectionSource interface {
	Kind() collection.SourceKind
	Observe(ctx context.Context) (collection.Observation, error)
}

// Catalog resolve identidades de carta; ausência não é erro.
type Catalog interface {
	ResolvePorArena(ctx context.Context, id collection.ArenaID) (collection.CardIdentity, bool, error)
	ResolvePorNome(ctx context.Context, nome string) (collection.CardIdentity, bool, error)
}

// SnapshotStore persiste snapshots imutáveis e expõe o corrente.
type SnapshotStore interface {
	Save(ctx context.Context, snap collection.Snapshot) error
	Get(ctx context.Context, id collection.SnapshotID) (collection.Snapshot, error)
	Latest(ctx context.Context) (collection.Snapshot, bool, error)
}

// DeckStore persiste revisões locais de deck com sua proveniência.
type DeckStore interface {
	SaveRevision(ctx context.Context, rev decks.Revision) error
	Revisions(ctx context.Context, deckName string) ([]decks.Revision, error)
}

// ExportFormat identifica os formatos de projeção de compatibilidade.
type ExportFormat string

// Formatos suportados pelo exportador de compatibilidade.
const (
	ExportJSON ExportFormat = "json"
	ExportCSV  ExportFormat = "csv"
	ExportText ExportFormat = "txt"
)

// Exporter projeta um snapshot no formato de compatibilidade pedido; a
// escrita em disco é um efeito separado, aprovado à parte.
type Exporter interface {
	Export(ctx context.Context, snap collection.Snapshot, formato ExportFormat) ([]byte, error)
}
