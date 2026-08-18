package testkit

import (
	"encoding/json"
	"maps"
	"testing"
)

func TestFixturesAreDeterministicAndSized(t *testing.T) {
	if len(FixtureSmall()) != 5 {
		t.Fatalf("small fixture must hold 5 entries: %d", len(FixtureSmall()))
	}
	large := FixtureLarge()
	if len(large) != 5000 {
		t.Fatalf("large fixture must hold 5000 entries: %d", len(large))
	}
	if !maps.Equal(large, FixtureLarge()) {
		t.Fatal("the large fixture must be deterministic")
	}
	db := FixtureCardDB()
	for grp := range large {
		if _, ok := db[grp]; !ok {
			t.Fatalf("large fixture references unknown grp %d", grp)
		}
	}
}

// TestDuplicatePrintingMergesByName: playset math counts reprints of
// the same logical card together.
func TestDuplicatePrintingMergesByName(t *testing.T) {
	db := FixtureCardDB()
	byName := map[string]int{}
	for grp, count := range FixtureDuplicatePrinting() {
		byName[db[grp].Name] += count
	}
	if byName["Synthetic Bolt"] != 4 {
		t.Fatalf("printings must merge to a full playset: %v", byName)
	}
	if db[1001].Set == db[1005].Set {
		t.Fatal("the duplicate printing must come from another set")
	}
}

func TestUnknownCardIsDetectable(t *testing.T) {
	db := FixtureCardDB()
	unknown := 0
	for grp := range FixtureUnknownCard() {
		if _, ok := db[grp]; !ok {
			unknown++
		}
	}
	if unknown != 1 {
		t.Fatalf("exactly one unknown card must be present: %d", unknown)
	}
}

// TestNearlyBuildableDeckIsOneCopyShort: counting reprints together,
// the deck misses exactly one copy of one card.
func TestNearlyBuildableDeckIsOneCopyShort(t *testing.T) {
	deck, collection := FixtureNearlyBuildableDeck()
	db := FixtureCardDB()
	owned := map[string]int{}
	for grp, count := range collection {
		owned[db[grp].Name] += count
	}
	missing := 0
	for _, want := range deck {
		if short := want.Count - owned[db[want.GrpID].Name]; short > 0 {
			missing += short
		}
	}
	if missing != 1 {
		t.Fatalf("the deck must be exactly one copy short: %d", missing)
	}
}

func TestMalformedFixtureFailsParsing(t *testing.T) {
	var out map[string]any
	if err := json.Unmarshal(FixtureMalformedCollection(), &out); err == nil {
		t.Fatal("the malformed fixture must fail any parser")
	}
}
