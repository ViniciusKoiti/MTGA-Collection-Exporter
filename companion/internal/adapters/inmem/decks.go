package inmem

import (
	"context"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// DeckStore guarda revisões de deck em memória, por nome de deck.
type DeckStore struct {
	mu      sync.Mutex
	porDeck map[string][]decks.Revision
}

var _ ports.DeckStore = (*DeckStore)(nil)

// NewDeckStore cria o store vazio.
func NewDeckStore() *DeckStore {
	return &DeckStore{porDeck: make(map[string][]decks.Revision)}
}

// SaveRevision anexa a revisão à história do deck.
func (d *DeckStore) SaveRevision(_ context.Context, rev decks.Revision) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.porDeck[rev.Deck.Name] = append(d.porDeck[rev.Deck.Name], rev)
	return nil
}

// Revisions devolve as revisões na ordem em que foram salvas.
func (d *DeckStore) Revisions(_ context.Context, deckName string) ([]decks.Revision, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]decks.Revision(nil), d.porDeck[deckName]...), nil
}
