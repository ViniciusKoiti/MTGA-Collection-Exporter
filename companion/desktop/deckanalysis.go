//go:build windows

package main

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// contextOrBackground guards bindings called before startup.
func (a *App) contextOrBackground() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

// DeckAnalysis is everything the Decks workspace renders for one
// deck text: normalized preview, verdict, ownership and state.
type DeckAnalysis struct {
	Preview   string          `json:"preview"`
	Verdict   decks.Veredicto `json:"verdict"`
	Ownership decks.Ownership `json:"ownership"`
	State     viewstate.State `json:"state"`
}

// AnalyzeDeck parses the Arena text, validates it against the
// workspace ruleset and compares ownership with the latest snapshot.
func (a *App) AnalyzeDeck(name, text string) DeckAnalysis {
	deck, err := parseDeck(name, text)
	if err != nil {
		return DeckAnalysis{State: viewstate.FromError(err)}
	}
	analysis := DeckAnalysis{Preview: deck.ArenaText(),
		Verdict: workspaceRuleset().Validate(deck, systemClock{}.Now())}
	reader, err := collections()
	if err != nil {
		analysis.State = viewstate.FromError(err)
		return analysis
	}
	snap, ok, err := reader.Latest(a.contextOrBackground())
	if err != nil {
		analysis.State = viewstate.FromError(err)
		return analysis
	}
	if !ok {
		analysis.State = viewstate.Empty() // ownership needs a snapshot
		return analysis
	}
	analysis.Ownership = decks.CompareOwnership(deck, snap)
	analysis.State = viewstate.Success()
	return analysis
}

// DeckRevision is one history line of the workspace; the verdict is
// reproducible from the recorded snapshot + ruleset identity.
type DeckRevision struct {
	ID      string `json:"id"`
	SavedAt string `json:"saved_at"`
	Ruleset string `json:"ruleset"`
}

// DeckHistory lists the saved revisions of a deck, newest last.
func (a *App) DeckHistory(name string) []DeckRevision {
	service, err := deckService()
	if err != nil {
		return nil
	}
	revisions, err := service.History(a.contextOrBackground(), name)
	if err != nil {
		return nil
	}
	out := make([]DeckRevision, 0, len(revisions))
	for _, revision := range revisions {
		out = append(out, DeckRevision{ID: revision.ID,
			SavedAt: revision.SavedAt.UTC().Format("2006-01-02 15:04"),
			Ruleset: fmt.Sprintf("%s v%d", revision.Ruleset.Format,
				revision.Ruleset.Version)})
	}
	return out
}
