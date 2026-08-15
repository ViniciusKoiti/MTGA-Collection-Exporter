package decks

import (
	"sort"
	"strings"
)

// CardProfile carries the catalog metadata a substitution decision needs:
// colors, mana value, type line, and format legality tags. It is injected
// from the catalog adapter; the domain never fetches data on its own.
// (OpenSpec introduce-agentic-go-companion, task 5.4.)
type CardProfile struct {
	Name      string
	Colors    []string // e.g. ["R"]
	ManaValue int
	Types     []string // e.g. ["Instant"]
	Formats   []string // formats where the card is legal
}

// SubstitutionCandidate is one ranked suggestion with documented evidence:
// every scored dimension is listed so the recommendation is auditable.
type SubstitutionCandidate struct {
	Name     string
	Score    int
	Evidence []string // stable reason codes: type_match, color_match, ...
}

// Scoring weights, documented as part of the ranking evidence contract.
const (
	scoreTypeMatch  = 4
	scoreColorMatch = 3
	scoreManaExact  = 2
	scoreManaClose  = 1
)

// SubstitutionCandidates ranks owned cards that could replace a missing
// card. Candidates must be legal in the requested format; ranking is
// deterministic (score descending, then name ascending) and every score
// component is recorded as evidence.
func SubstitutionCandidates(
	missing CardProfile,
	owned []CardProfile,
	format string,
	max int,
) []SubstitutionCandidate {
	var candidates []SubstitutionCandidate
	for _, card := range owned {
		if strings.EqualFold(card.Name, missing.Name) || !legalIn(card, format) {
			continue
		}
		candidate := SubstitutionCandidate{Name: card.Name,
			Evidence: []string{"owned", "format_legal:" + format}}
		if overlaps(card.Types, missing.Types) {
			candidate.Score += scoreTypeMatch
			candidate.Evidence = append(candidate.Evidence, "type_match")
		}
		if sameSet(card.Colors, missing.Colors) {
			candidate.Score += scoreColorMatch
			candidate.Evidence = append(candidate.Evidence, "color_match")
		}
		switch delta := card.ManaValue - missing.ManaValue; {
		case delta == 0:
			candidate.Score += scoreManaExact
			candidate.Evidence = append(candidate.Evidence, "mana_match")
		case delta == 1 || delta == -1:
			candidate.Score += scoreManaClose
			candidate.Evidence = append(candidate.Evidence, "mana_close")
		}
		candidates = append(candidates, candidate)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Name < candidates[j].Name
	})
	if max > 0 && len(candidates) > max {
		candidates = candidates[:max]
	}
	return candidates
}
