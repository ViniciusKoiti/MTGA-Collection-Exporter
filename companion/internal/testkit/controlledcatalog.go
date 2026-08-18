package testkit

import (
	"context"
	"fmt"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// ControlledCatalog is the provider adapter behind ports.Catalog: it
// resolves identities from the sanitized fixture card database (task
// 5.6), so provider behavior is fully scenario-owned. Absence is not
// an error, exactly as the port demands.
type ControlledCatalog struct {
	DB map[int]CardMeta
}

var _ ports.Catalog = ControlledCatalog{}

func (c ControlledCatalog) identity(grp int, meta CardMeta) collection.CardIdentity {
	return collection.CardIdentity{
		Printing: collection.PrintingID(fmt.Sprintf("print-%d", grp)),
		Arena:    collection.ArenaID(grp),
		Oracle:   collection.OracleID("oracle-" + meta.Name),
		Name:     meta.Name,
		Set:      meta.Set,
	}
}

// ResolvePorArena resolves one arena ID; unknown IDs answer ok=false.
func (c ControlledCatalog) ResolvePorArena(_ context.Context,
	id collection.ArenaID) (collection.CardIdentity, bool, error) {
	meta, ok := c.DB[int(id)]
	if !ok {
		return collection.CardIdentity{}, false, nil
	}
	return c.identity(int(id), meta), true, nil
}

// ResolvePorNome resolves by name, picking the LOWEST grp id among
// printings so resolution stays deterministic.
func (c ControlledCatalog) ResolvePorNome(_ context.Context,
	nome string) (collection.CardIdentity, bool, error) {
	best := -1
	for grp, meta := range c.DB {
		if meta.Name == nome && (best == -1 || grp < best) {
			best = grp
		}
	}
	if best == -1 {
		return collection.CardIdentity{}, false, nil
	}
	return c.identity(best, c.DB[best]), true, nil
}

// ControlledClipboard is the local-effect adapter: it records every
// write so scenarios can assert effects without touching the OS.
type ControlledClipboard struct {
	mu     sync.Mutex
	Writes []string
}

var _ ports.Clipboard = (*ControlledClipboard)(nil)

// Write records the effect.
func (c *ControlledClipboard) Write(_ context.Context, texto string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Writes = append(c.Writes, texto)
	return nil
}
