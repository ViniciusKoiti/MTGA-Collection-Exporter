// Package decksvc is the application service for local saved-deck
// revisions (OpenSpec introduce-agentic-go-companion, task 5.5): every
// revision is linked to the collection snapshot and the exact ruleset
// version used for validation, so past verdicts stay reproducible.
package decksvc

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Service persists deck revisions with their validation provenance.
type Service struct {
	snapshots ports.SnapshotStore
	store     ports.DeckStore
	clock     ports.Clock
	nextID    int
}

// New wires the service; IDs are sequential and deterministic under the
// controlled clock profile.
func New(snapshots ports.SnapshotStore, store ports.DeckStore, clock ports.Clock) *Service {
	return &Service{snapshots: snapshots, store: store, clock: clock, nextID: 1}
}

// SaveRevision validates the deck against the ruleset and the current
// snapshot, then persists the revision regardless of legality — the
// verdict (with its ruleset identity) is part of the recorded history.
func (s *Service) SaveRevision(
	ctx context.Context,
	deck decks.Deck,
	ruleset decks.Standard,
) (decks.Revision, decks.Veredicto, error) {
	snap, exists, err := s.snapshots.Latest(ctx)
	if err != nil {
		return decks.Revision{}, decks.Veredicto{},
			apperr.New(apperr.CodeInternal, "decks.save_revision", err)
	}
	if !exists {
		return decks.Revision{}, decks.Veredicto{},
			apperr.New(apperr.CodeSnapshotNotFound, "decks.save_revision", nil)
	}
	now := s.clock.Now()
	verdict := ruleset.Validate(deck, now)
	revision, err := decks.NewRevision(
		fmt.Sprintf("rev-%06d", s.nextID), deck, snap.ID, verdict.Ruleset, now)
	if err != nil {
		return decks.Revision{}, verdict,
			apperr.New(apperr.CodeValidationFailed, "decks.save_revision", err)
	}
	if err := s.store.SaveRevision(ctx, revision); err != nil {
		return decks.Revision{}, verdict,
			apperr.New(apperr.CodeInternal, "decks.save_revision", err)
	}
	s.nextID++
	return revision, verdict, nil
}

// History returns the saved revisions of a deck in save order.
func (s *Service) History(ctx context.Context, deckName string) ([]decks.Revision, error) {
	revisions, err := s.store.Revisions(ctx, deckName)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "decks.history", err)
	}
	return revisions, nil
}
