package collectionsvc

import (
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func snapshotFixture() collection.Snapshot {
	entry := func(printing, name, set string, qty int) collection.Entry {
		return collection.Entry{Identity: collection.CardIdentity{
			Printing: collection.PrintingID(printing), Name: name, Set: set,
			Oracle: collection.OracleID("o-" + printing), Arena: 100},
			Quantity: qty}
	}
	return collection.Snapshot{Entries: []collection.Entry{
		entry("p1", "Lightning Strike", "DMU", 4),
		entry("p2", "Counterspell", "DMU", 2),
		entry("p3", "Lightning Bolt", "STA", 1),
		{Unresolved: true, Raw: "??? Lightning ???", Quantity: 1},
	}}
}

func TestCompoundFiltersCompose(t *testing.T) {
	rows := Page(snapshotFixture(), Query{Text: "lightning",
		Set: "dmu", MinQuantity: 2})
	if len(rows) != 1 || rows[0].Name != "Lightning Strike" {
		t.Fatalf("text+set+quantity must compose: %+v", rows)
	}
	unresolved := Page(snapshotFixture(), Query{UnresolvedOnly: true})
	if len(unresolved) != 1 || !unresolved[0].Unresolved ||
		unresolved[0].Raw == "" {
		t.Fatalf("unresolved records must filter and keep their raw: %+v",
			unresolved)
	}
}

func TestSortIsDeterministicWithStableFallbacks(t *testing.T) {
	byQuantity := Page(snapshotFixture(), Query{Sort: "quantity"})
	if byQuantity[0].Quantity != 4 || byQuantity[len(byQuantity)-1].Quantity != 1 {
		t.Fatalf("quantity sorts descending: %+v", byQuantity)
	}
	bySet := Page(snapshotFixture(), Query{Sort: "set"})
	if bySet[0].Set > bySet[len(bySet)-1].Set {
		t.Fatalf("set sorts ascending: %+v", bySet)
	}
	unknown := Page(snapshotFixture(), Query{Sort: "mana-symbols"})
	if unknown[0].Name > unknown[1].Name {
		t.Fatalf("unknown sort keys must fall back to name: %+v", unknown)
	}
}

func TestRowsCarryTheDetailFields(t *testing.T) {
	rows := Page(snapshotFixture(), Query{Text: "strike"})
	if len(rows) != 1 || rows[0].Oracle != "o-p1" || rows[0].Arena != 100 ||
		rows[0].Printing != "p1" {
		t.Fatalf("details must ride along: %+v", rows)
	}
}
