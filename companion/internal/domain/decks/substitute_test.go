package decks

import (
	"strings"
	"testing"
)

func missingCard() CardProfile {
	return CardProfile{Name: "Lightning Strike", Colors: []string{"R"},
		ManaValue: 2, Types: []string{"Instant"}, Formats: []string{"standard"}}
}

func ownedPool() []CardProfile {
	return []CardProfile{
		{Name: "Play with Fire", Colors: []string{"R"}, ManaValue: 1,
			Types: []string{"Instant"}, Formats: []string{"standard"}},
		{Name: "Shock", Colors: []string{"R"}, ManaValue: 1,
			Types: []string{"Instant"}, Formats: []string{"standard"}},
		{Name: "Abrade", Colors: []string{"R"}, ManaValue: 2,
			Types: []string{"Instant"}, Formats: []string{"standard"}},
		{Name: "Legacy Bolt", Colors: []string{"R"}, ManaValue: 1,
			Types: []string{"Instant"}, Formats: []string{"legacy"}}, // wrong format
		{Name: "Grizzly Bears", Colors: []string{"G"}, ManaValue: 2,
			Types: []string{"Creature"}, Formats: []string{"standard"}},
		{Name: "Lightning Strike", Colors: []string{"R"}, ManaValue: 2,
			Types: []string{"Instant"}, Formats: []string{"standard"}}, // the card itself
	}
}

func TestSubstitutionRankingIsDeterministicWithEvidence(t *testing.T) {
	candidates := SubstitutionCandidates(missingCard(), ownedPool(), "standard", 0)
	names := make([]string, len(candidates))
	for i, c := range candidates {
		names[i] = c.Name
	}
	// Abrade: type+color+mana exact (9). Play with Fire / Shock: type+color+
	// mana close (8), tie broken by name. Grizzly Bears: mana exact only (2).
	want := []string{"Abrade", "Play with Fire", "Shock", "Grizzly Bears"}
	for i, name := range want {
		if names[i] != name {
			t.Fatalf("ranking out of order: %v", names)
		}
	}
	best := candidates[0]
	if best.Score != 9 {
		t.Fatalf("unexpected top score: %+v", best)
	}
	evidence := strings.Join(best.Evidence, ",")
	for _, reason := range []string{"owned", "format_legal:standard",
		"type_match", "color_match", "mana_match"} {
		if !strings.Contains(evidence, reason) {
			t.Fatalf("missing evidence %q: %+v", reason, best.Evidence)
		}
	}
}

func TestSubstitutionFiltersFormatSelfAndLimit(t *testing.T) {
	candidates := SubstitutionCandidates(missingCard(), ownedPool(), "standard", 2)
	if len(candidates) != 2 {
		t.Fatalf("limit ignored: %+v", candidates)
	}
	for _, c := range candidates {
		if c.Name == "Legacy Bolt" {
			t.Fatal("format-illegal card must be filtered out")
		}
		if c.Name == "Lightning Strike" {
			t.Fatal("the missing card itself is not a substitute")
		}
	}
	if got := SubstitutionCandidates(missingCard(), nil, "standard", 0); len(got) != 0 {
		t.Fatalf("empty pool yields no candidates: %+v", got)
	}
}
