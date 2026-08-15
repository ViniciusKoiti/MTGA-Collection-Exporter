package decksvc

import (
	"errors"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

func fixture(t *testing.T, seeded bool) (*Service, *inmem.Clock) {
	t.Helper()
	clock := inmem.NewClock(time.Unix(1_700_000_000, 0))
	snapshots := inmem.NewSnapshotStore()
	if seeded {
		obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
			clock.Now().Add(-time.Hour), nil, nil)
		snap, err := collection.NewSnapshot("snap-1", obs, clock.Now(),
			[]collection.Entry{{Identity: collection.CardIdentity{
				Printing: "p1", Arena: 1, Name: "Mountain", Set: "UNF"}, Quantity: 20}}, nil)
		if err != nil {
			t.Fatalf("snapshot: %v", err)
		}
		if err := snapshots.Save(t.Context(), snap); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return New(snapshots, inmem.NewDeckStore(), clock), clock
}

func standardDeck(t *testing.T) decks.Deck {
	t.Helper()
	deck, err := decks.NewDeck("Mono Red", []decks.Entry{
		{Name: "Lightning Strike", Quantity: 4},
		{Name: "Mountain", Quantity: 56},
	}, nil)
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	return deck
}

func TestRevisionLinksSnapshotAndRulesetVersion(t *testing.T) {
	service, clock := fixture(t, true)
	ruleset := decks.StandardPadrao(map[string]bool{"lightning strike": true}, clock.Now())
	revision, verdict, err := service.SaveRevision(t.Context(), standardDeck(t), ruleset)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if revision.Snapshot != "snap-1" || revision.Ruleset.String() != "standard/v1" ||
		!verdict.Legal || revision.ID != "rev-000001" {
		t.Fatalf("provenance incomplete: %+v / %+v", revision, verdict)
	}
	clock.Advance(time.Minute)
	if _, _, err := service.SaveRevision(t.Context(), standardDeck(t), ruleset); err != nil {
		t.Fatalf("second save: %v", err)
	}
	history, err := service.History(t.Context(), "Mono Red")
	if err != nil || len(history) != 2 || history[1].ID != "rev-000002" ||
		!history[1].SavedAt.After(history[0].SavedAt) {
		t.Fatalf("history out of order: %+v (%v)", history, err)
	}
}

func TestIllegalDeckIsStillRecordedWithItsVerdict(t *testing.T) {
	service, clock := fixture(t, true)
	ruleset := decks.StandardPadrao(map[string]bool{}, clock.Now()) // nothing legal
	revision, verdict, err := service.SaveRevision(t.Context(), standardDeck(t), ruleset)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if verdict.Legal || len(verdict.Problemas) == 0 || revision.Ruleset != verdict.Ruleset {
		t.Fatalf("verdict should record the violations: %+v", verdict)
	}
}

func TestSaveWithoutSnapshotFailsWithStableCode(t *testing.T) {
	service, clock := fixture(t, false)
	ruleset := decks.StandardPadrao(nil, clock.Now())
	_, _, err := service.SaveRevision(t.Context(), standardDeck(t), ruleset)
	if apperr.CodeOf(err) != apperr.CodeSnapshotNotFound {
		t.Fatalf("expected snapshot_not_found, got: %v", err)
	}
	var typed *apperr.Error
	if !errors.As(err, &typed) {
		t.Fatal("error should be the typed application error")
	}
}
