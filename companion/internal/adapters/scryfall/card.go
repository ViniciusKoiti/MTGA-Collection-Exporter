// Package scryfall caches the Scryfall bulk card data locally (OpenSpec
// introduce-agentic-go-companion, task 3.6): explicit version and
// freshness metadata, bounded retries, request timeouts injected by the
// composition, and an offline fallback that serves the last good cache
// flagged as stale. Tests exercise everything against httptest servers —
// the suite never touches the real API.
package scryfall

import (
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// bulkIndex is the minimal shape of GET /bulk-data.
type bulkIndex struct {
	Data []bulkEntry `json:"data"`
}

type bulkEntry struct {
	Type        string `json:"type"`
	UpdatedAt   string `json:"updated_at"`
	DownloadURI string `json:"download_uri"`
}

// bulkCard is the minimal per-card shape of the default_cards bulk
// file, plus the metadata substitutions need (companion task 5.6).
type bulkCard struct {
	ID              string            `json:"id"`
	OracleID        string            `json:"oracle_id"`
	Name            string            `json:"name"`
	Set             string            `json:"set"`
	CollectorNumber string            `json:"collector_number"`
	ArenaID         int               `json:"arena_id"`
	Colors          []string          `json:"colors"`
	Cmc             float64           `json:"cmc"`
	TypeLine        string            `json:"type_line"`
	Legalities      map[string]string `json:"legalities"`
}

// identity maps a bulk card to the domain identity; cards without an
// arena id are kept out of the arena index but remain name-resolvable.
func (c bulkCard) identity() collection.CardIdentity {
	return collection.CardIdentity{
		Printing: collection.PrintingID(c.ID),
		Arena:    collection.ArenaID(c.ArenaID),
		Oracle:   collection.OracleID(c.OracleID),
		Name:     c.Name,
		Set:      c.Set,
	}
}
