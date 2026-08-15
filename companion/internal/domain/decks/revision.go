package decks

import (
	"fmt"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Revision é uma revisão local salva de um deck, sempre ligada ao snapshot
// de coleção e à versão de ruleset usados na validação (tarefa 5.5 do
// companion; modelo exigido pelo DeckStore).
type Revision struct {
	ID       string
	Deck     Deck
	Snapshot collection.SnapshotID
	Ruleset  RulesetID
	SavedAt  time.Time
}

// NewRevision valida os invariantes da revisão.
func NewRevision(
	id string,
	deck Deck,
	snapshot collection.SnapshotID,
	ruleset RulesetID,
	savedAt time.Time,
) (Revision, error) {
	if id == "" {
		return Revision{}, fmt.Errorf("decks: revisão sem ID")
	}
	if snapshot == "" {
		return Revision{}, fmt.Errorf("decks: revisão %s sem snapshot de referência", id)
	}
	if err := ruleset.Valida(); err != nil {
		return Revision{}, err
	}
	if savedAt.IsZero() {
		return Revision{}, fmt.Errorf("decks: revisão %s sem instante", id)
	}
	return Revision{ID: id, Deck: deck, Snapshot: snapshot, Ruleset: ruleset, SavedAt: savedAt}, nil
}
