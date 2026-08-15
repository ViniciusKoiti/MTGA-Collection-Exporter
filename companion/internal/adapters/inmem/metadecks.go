package inmem

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// MetaDecks devolve um catálogo fixo de meta decks para o harness.
type MetaDecks struct {
	Itens []ports.MetaDeck
}

var _ ports.MetaDeckCatalog = (*MetaDecks)(nil)

// Decks filtra o catálogo fixo pelo formato pedido.
func (m *MetaDecks) Decks(_ context.Context, formato string) ([]ports.MetaDeck, error) {
	var filtrados []ports.MetaDeck
	for _, deck := range m.Itens {
		if deck.Formato == formato {
			filtrados = append(filtrados, deck)
		}
	}
	return filtrados, nil
}
