//go:build windows

package main

import (
	"os"
	"path/filepath"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/scryfall"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// Substitution is the ranked suggestions for one missing card.
type Substitution struct {
	Missing    string                        `json:"missing"`
	Candidates []decks.SubstitutionCandidate `json:"candidates"`
}

// SubstitutionReport is what the workspace renders for one deck.
type SubstitutionReport struct {
	Substitutions []Substitution  `json:"substitutions"`
	State         viewstate.State `json:"state"`
}

func scryfallDir() (string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(config, "MTGA-Companion", "scryfall"), nil
}

// SuggestSubstitutions ranks owned replacements for the deck's
// missing cards using the cached catalog profiles — cache-only, so
// this can never start a bulk download mid-click.
func (a *App) SuggestSubstitutions(name, text string) SubstitutionReport {
	deck, err := parseDeck(name, text)
	if err != nil {
		return SubstitutionReport{State: viewstate.FromError(err)}
	}
	reader, err := collections()
	if err != nil {
		return SubstitutionReport{State: viewstate.FromError(err)}
	}
	snap, ok, err := reader.Latest(a.contextOrBackground())
	if err != nil {
		return SubstitutionReport{State: viewstate.FromError(err)}
	}
	if !ok {
		return SubstitutionReport{State: viewstate.Empty()}
	}
	dir, err := scryfallDir()
	if err != nil {
		return SubstitutionReport{State: viewstate.FromError(err)}
	}
	catalog, err := scryfall.OpenCachedOnly(dir)
	if err != nil {
		return SubstitutionReport{State: viewstate.Partial(
			apperr.CodeCatalogUnavailable,
			"the card catalog is not downloaded yet")}
	}
	var pool []decks.CardProfile
	for _, entry := range snap.Entries {
		if profile, has := catalog.Profile(entry.Identity.Name); has {
			pool = append(pool, profile)
		}
	}
	report := SubstitutionReport{State: viewstate.Stale(
		"catalog freshness unknown (cached copy)")}
	for _, line := range decks.CompareOwnership(deck, snap).Lines {
		if line.Missing == 0 {
			continue
		}
		missing, has := catalog.Profile(line.Name)
		if !has {
			continue // no profile, no invented suggestion
		}
		report.Substitutions = append(report.Substitutions, Substitution{
			Missing: line.Name,
			Candidates: decks.SubstitutionCandidates(missing, pool,
				"standard", 3)})
	}
	return report
}
