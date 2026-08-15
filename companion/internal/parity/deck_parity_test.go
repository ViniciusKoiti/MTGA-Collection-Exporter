package parity

import (
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

func TestDeckCheckMatchesThePythonMCP(t *testing.T) {
	var pyDeck struct {
		Buildable    bool `json:"buildable"`
		MissingCount int  `json:"missing_count"`
		Missing      []struct {
			Name string `json:"name"`
			Need int    `json:"need"`
		} `json:"missing"`
	}
	golden(t, "python_check_deck.golden.json", &pyDeck)
	deck, err := decks.NewDeck("parity", []decks.Entry{
		{Name: "Lightning Strike", Quantity: 4},
		{Name: "Carta Inexistente", Quantity: 2},
	}, nil)
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	ownership := decks.CompareOwnership(deck, sharedSnapshot(t))
	if ownership.Complete != pyDeck.Buildable {
		t.Fatalf("buildability diverged: Go %t vs Python %t",
			ownership.Complete, pyDeck.Buildable)
	}
	if ownership.TotalMissing != pyDeck.MissingCount {
		t.Fatalf("missing totals diverged: Go %d vs Python %d",
			ownership.TotalMissing, pyDeck.MissingCount)
	}
	for _, missing := range pyDeck.Missing {
		matched := false
		for _, line := range ownership.Lines {
			if line.Name == missing.Name && line.Missing == missing.Need {
				matched = true
			}
		}
		if !matched {
			t.Fatalf("Python misses %q x%d; Go disagrees: %+v",
				missing.Name, missing.Need, ownership.Lines)
		}
	}
}
