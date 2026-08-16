package collectionsvc

import (
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func snapWith(entries ...collection.Entry) collection.Snapshot {
	return collection.Snapshot{Entries: entries}
}

func card(printing, name string, qty int) collection.Entry {
	return collection.Entry{Identity: collection.CardIdentity{
		Printing: collection.PrintingID(printing), Name: name, Set: "TST"},
		Quantity: qty}
}

// TestCompareFindsAddedRemovedAndChangedDeterministically: the diff
// carries totals and one change row per differing card, sorted.
func TestCompareFindsAddedRemovedAndChangedDeterministically(t *testing.T) {
	previous := snapWith(card("p1", "Bolt", 3), card("p2", "Counter", 2),
		card("p3", "Giant", 1))
	latest := snapWith(card("p1", "Bolt", 4), card("p2", "Counter", 2),
		card("p4", "Angel", 2))
	diff := Compare(previous, latest)
	if diff.TotalBefore != 6 || diff.TotalAfter != 8 {
		t.Fatalf("totals must ride along: %+v", diff)
	}
	if len(diff.Changes) != 3 {
		t.Fatalf("expected added+removed+changed: %+v", diff.Changes)
	}
	if diff.Changes[0].Name != "Angel" || diff.Changes[0].Before != 0 ||
		diff.Changes[0].After != 2 {
		t.Fatalf("added must show 0 -> n: %+v", diff.Changes[0])
	}
	if diff.Changes[1].Name != "Bolt" || diff.Changes[1].Before != 3 ||
		diff.Changes[1].After != 4 {
		t.Fatalf("changed must show both counts: %+v", diff.Changes[1])
	}
	if diff.Changes[2].Name != "Giant" || diff.Changes[2].After != 0 {
		t.Fatalf("removed must show n -> 0: %+v", diff.Changes[2])
	}
}

func TestCompareKeysUnresolvedRecordsByTheirRawText(t *testing.T) {
	previous := snapWith(collection.Entry{Unresolved: true,
		Raw: "??? mystery", Quantity: 1})
	latest := snapWith(collection.Entry{Unresolved: true,
		Raw: "??? mystery", Quantity: 3})
	diff := Compare(previous, latest)
	if len(diff.Changes) != 1 || diff.Changes[0].Key != "raw:??? mystery" ||
		diff.Changes[0].Before != 1 || diff.Changes[0].After != 3 {
		t.Fatalf("unresolved records must diff by raw key: %+v", diff.Changes)
	}
	if identical := Compare(latest, latest); len(identical.Changes) != 0 {
		t.Fatalf("identical snapshots must diff empty: %+v", identical.Changes)
	}
}
