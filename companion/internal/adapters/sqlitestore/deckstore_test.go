package sqlitestore

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

func revisionFixture(t *testing.T, id string, at time.Time) decks.Revision {
	t.Helper()
	deck, err := decks.NewDeck("Mono Red", []decks.Entry{
		{Name: "Lightning Strike", Quantity: 4, Set: "DMU", Numero: "123"},
		{Name: "Mountain", Quantity: 20}}, nil)
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	rev, err := decks.NewRevision(id, deck, "snap-1",
		decks.RulesetID{Format: "standard", Version: 1}, at)
	if err != nil {
		t.Fatalf("revision: %v", err)
	}
	return rev
}

// TestDeckRevisionsSurviveRestartInOrderAndImmutable: history outlives
// a reopen, keeps save order, round-trips the deck, and a duplicate
// revision ID is refused — revisions never overwrite.
func TestDeckRevisionsSurviveRestartInOrderAndImmutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "decks.db")
	first, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	ctx := t.Context()
	base := time.Unix(1_700_000_000, 0).UTC()
	if err := first.Decks().SaveRevision(ctx,
		revisionFixture(t, "rev-000001", base)); err != nil {
		t.Fatalf("save 1: %v", err)
	}
	if err := first.Decks().SaveRevision(ctx,
		revisionFixture(t, "rev-000002", base.Add(time.Hour))); err != nil {
		t.Fatalf("save 2: %v", err)
	}
	if err := first.Decks().SaveRevision(ctx,
		revisionFixture(t, "rev-000001", base)); err == nil {
		t.Fatal("revisions are immutable: a duplicate ID must be refused")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	revisions, err := reopened.Decks().Revisions(ctx, "Mono Red")
	if err != nil || len(revisions) != 2 {
		t.Fatalf("history must survive the restart: %d %v",
			len(revisions), err)
	}
	if revisions[0].ID != "rev-000001" || revisions[1].ID != "rev-000002" {
		t.Fatalf("save order must hold: %+v", revisions)
	}
	loaded := revisions[0]
	if loaded.Deck.Name != "Mono Red" || len(loaded.Deck.Main) != 2 ||
		loaded.Deck.Main[0].Numero != "123" ||
		loaded.Ruleset != (decks.RulesetID{Format: "standard", Version: 1}) ||
		!loaded.SavedAt.Equal(base) || loaded.Snapshot != "snap-1" {
		t.Fatalf("the revision must round-trip completely: %+v", loaded)
	}
}
