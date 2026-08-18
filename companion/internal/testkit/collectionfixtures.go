package testkit

import "fmt"

// Collection fixtures (OpenSpec add-graph-workflow-harness, task 5.6).
// Every fixture is sanitized by construction: synthetic grp_ids and
// names, no real account data. The large fixture is generated
// deterministically so the repository never carries bulk data.

// CardMeta is the minimal card-database row scenarios need.
type CardMeta struct {
	Name string
	Set  string
}

// FixtureCardDB is the sanitized card database backing the fixtures.
func FixtureCardDB() map[int]CardMeta {
	db := map[int]CardMeta{
		1001: {Name: "Synthetic Bolt", Set: "TST"},
		1002: {Name: "Synthetic Counter", Set: "TST"},
		1003: {Name: "Synthetic Giant", Set: "TST"},
		1004: {Name: "Synthetic Angel", Set: "TST"},
		1005: {Name: "Synthetic Bolt", Set: "RPR"}, // reprint of 1001
	}
	for grp := 2000; grp < 7000; grp++ {
		db[grp] = CardMeta{Name: fmt.Sprintf("Bulk Card %d", grp), Set: "BLK"}
	}
	return db
}

// FixtureSmall is a five-entry collection.
func FixtureSmall() map[int]int {
	return map[int]int{1001: 4, 1002: 2, 1003: 1, 1004: 3, 1005: 1}
}

// FixtureLarge is a deterministic 5000-entry collection.
func FixtureLarge() map[int]int {
	cards := make(map[int]int, 5000)
	for grp := 2000; grp < 7000; grp++ {
		cards[grp] = grp%4 + 1
	}
	return cards
}

// FixtureDuplicatePrinting owns the same logical card under two
// printings: playset math must merge 1001 and 1005 by name.
func FixtureDuplicatePrinting() map[int]int {
	return map[int]int{1001: 3, 1005: 1}
}

// FixtureUnknownCard contains a grp_id absent from the card database.
func FixtureUnknownCard() map[int]int {
	return map[int]int{1001: 2, 999999: 1}
}

// DeckWant is one deck requirement line.
type DeckWant struct {
	GrpID int
	Count int
}

// FixtureNearlyBuildableDeck wants a playset the collection is one
// copy short of, counting reprints together.
func FixtureNearlyBuildableDeck() ([]DeckWant, map[int]int) {
	deck := []DeckWant{{GrpID: 1001, Count: 4}, {GrpID: 1002, Count: 2}}
	collection := map[int]int{1001: 2, 1005: 1, 1002: 2}
	return deck, collection
}

// FixtureMalformedCollection is raw bytes that must fail any parser.
func FixtureMalformedCollection() []byte {
	return []byte(`{"cards": {"1001": 4, "1002": `)
}
