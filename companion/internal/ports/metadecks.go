package ports

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// MetaDeck é um deck de meta publicado por um provedor aprovado, com a
// proveniência registrada para a evidência da recomendação.
type MetaDeck struct {
	Nome    string
	Formato string
	Fonte   string // provedor/snapshot de origem (proveniência)
	Deck    decks.Deck
}

// MetaDeckCatalog lista os meta decks do formato pedido; catálogo vazio
// não é erro — a decisão é do grafo.
type MetaDeckCatalog interface {
	Decks(ctx context.Context, formato string) ([]MetaDeck, error)
}
