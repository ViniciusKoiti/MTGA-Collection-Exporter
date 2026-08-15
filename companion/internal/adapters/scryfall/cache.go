package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

const (
	cardsFile = "scryfall_cards.json"
	metaFile  = "scryfall_meta.json"
)

// Catalog resolves identities from the cached bulk data.
type Catalog struct {
	byArena map[collection.ArenaID]collection.CardIdentity
	byName  map[string]collection.CardIdentity
	Meta    Metadata
	Stale   bool // true when served by the offline fallback
}

var _ ports.Catalog = (*Catalog)(nil)

// ResolvePorArena resolves by MTGA grp id; absence is not an error.
func (c *Catalog) ResolvePorArena(_ context.Context, id collection.ArenaID) (collection.CardIdentity, bool, error) {
	identity, ok := c.byArena[id]
	return identity, ok, nil
}

// ResolvePorNome resolves by case-insensitive exact name.
func (c *Catalog) ResolvePorNome(_ context.Context, name string) (collection.CardIdentity, bool, error) {
	identity, ok := c.byName[strings.ToLower(name)]
	return identity, ok, nil
}

// storeCache atomically writes the cards and metadata files.
func storeCache(dir string, cards []bulkCard, meta Metadata) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, value := range map[string]any{cardsFile: cards, metaFile: meta} {
		payload, err := json.Marshal(value)
		if err != nil {
			return err
		}
		temp := filepath.Join(dir, name+".tmp")
		if err := os.WriteFile(temp, payload, 0o644); err != nil {
			return err
		}
		if err := os.Rename(temp, filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// loadCache builds the catalog from the local cache files.
func loadCache(dir string) (*Catalog, error) {
	metaPayload, err := os.ReadFile(filepath.Join(dir, metaFile))
	if err != nil {
		return nil, fmt.Errorf("scryfall: no local cache: %w", err)
	}
	var meta Metadata
	if err := json.Unmarshal(metaPayload, &meta); err != nil {
		return nil, fmt.Errorf("scryfall: corrupted cache metadata: %w", err)
	}
	cardsPayload, err := os.ReadFile(filepath.Join(dir, cardsFile))
	if err != nil {
		return nil, err
	}
	var cards []bulkCard
	if err := json.Unmarshal(cardsPayload, &cards); err != nil {
		return nil, fmt.Errorf("scryfall: corrupted cache: %w", err)
	}
	catalog := &Catalog{Meta: meta,
		byArena: make(map[collection.ArenaID]collection.CardIdentity, len(cards)),
		byName:  make(map[string]collection.CardIdentity, len(cards))}
	for _, card := range cards {
		identity := card.identity()
		if card.ArenaID != 0 {
			catalog.byArena[identity.Arena] = identity
		}
		catalog.byName[strings.ToLower(card.Name)] = identity
	}
	return catalog, nil
}
