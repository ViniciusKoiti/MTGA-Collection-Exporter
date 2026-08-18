package sqlitestore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// DeckStore persists deck revisions durably (companion task 5.6).
type DeckStore struct {
	store *Store
}

var _ ports.DeckStore = (*DeckStore)(nil)

// Decks exposes the durable deck-revision store.
func (s *Store) Decks() *DeckStore { return &DeckStore{store: s} }

// SaveRevision appends one revision; revisions are immutable, so a
// duplicate ID is an error, never an overwrite.
func (d *DeckStore) SaveRevision(ctx context.Context, rev decks.Revision) error {
	deckJSON, err := json.Marshal(rev.Deck)
	if err != nil {
		return fmt.Errorf("sqlitestore: deck encode: %w", err)
	}
	ruleset := fmt.Sprintf("%s:%d", rev.Ruleset.Format, rev.Ruleset.Version)
	_, err = d.store.db.ExecContext(ctx, `INSERT INTO deck_revisions
		(id, deck_name, deck_json, snapshot_id, ruleset, saved_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rev.ID, rev.Deck.Name, string(deckJSON), string(rev.Snapshot),
		ruleset, formataInstante(rev.SavedAt))
	if err != nil {
		return fmt.Errorf("sqlitestore: deck revision save: %w", err)
	}
	return nil
}

// Revisions lists the saved revisions of a deck in save order.
func (d *DeckStore) Revisions(ctx context.Context,
	deckName string) ([]decks.Revision, error) {
	rows, err := d.store.db.QueryContext(ctx, `SELECT id, deck_json,
		snapshot_id, ruleset, saved_at FROM deck_revisions
		WHERE deck_name = ? ORDER BY saved_at, id`, deckName)
	if err != nil {
		return nil, fmt.Errorf("sqlitestore: deck revisions: %w", err)
	}
	defer rows.Close()
	var out []decks.Revision
	for rows.Next() {
		var rev decks.Revision
		var deckJSON, snapshot, ruleset, savedAt string
		if err := rows.Scan(&rev.ID, &deckJSON, &snapshot, &ruleset,
			&savedAt); err != nil {
			return nil, fmt.Errorf("sqlitestore: deck scan: %w", err)
		}
		if err := json.Unmarshal([]byte(deckJSON), &rev.Deck); err != nil {
			return nil, fmt.Errorf("sqlitestore: deck decode: %w", err)
		}
		rev.Snapshot = collection.SnapshotID(snapshot)
		if rev.Ruleset, err = parseRuleset(ruleset); err != nil {
			return nil, err
		}
		if rev.SavedAt, err = parseInstante(savedAt); err != nil {
			return nil, fmt.Errorf("sqlitestore: deck saved_at: %w", err)
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}

func parseRuleset(text string) (decks.RulesetID, error) {
	format, versionText, ok := strings.Cut(text, ":")
	if !ok {
		return decks.RulesetID{}, fmt.Errorf(
			"sqlitestore: malformed ruleset %q", text)
	}
	version, err := strconv.Atoi(versionText)
	if err != nil {
		return decks.RulesetID{}, fmt.Errorf(
			"sqlitestore: malformed ruleset version %q", text)
	}
	return decks.RulesetID{Format: format, Version: version}, nil
}
