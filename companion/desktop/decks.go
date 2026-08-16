//go:build windows

package main

import (
	"sync"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/decksvc"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

var (
	deckOnce sync.Once
	deckSvc  *decksvc.Service
)

// deckService wires revisions over the local snapshot store; the
// revision store itself is in-memory for now (session history) — the
// durable adapter is a follow-up, the port already exists.
func deckService() (*decksvc.Service, error) {
	var err error
	deckOnce.Do(func() {
		snapshots, openErr := collectionsStore()
		if openErr != nil {
			err = openErr
			return
		}
		deckSvc = decksvc.New(snapshots, inmem.NewDeckStore(), systemClock{})
	})
	if deckSvc == nil && err == nil {
		err = storeErr
	}
	return deckSvc, err
}

// workspaceRuleset builds the Standard ruleset with an EMPTY legality
// catalog and a zero timestamp: the verdict honestly reports
// CatalogoStale instead of inventing legality (the catalog arrives
// with the meta tasks).
func workspaceRuleset() decks.Standard {
	return decks.StandardPadrao(map[string]bool{}, time.Time{})
}

// parseDeck maps parse failures to the stable validation code.
func parseDeck(name, text string) (decks.Deck, error) {
	if name == "" {
		name = "Untitled"
	}
	deck, err := decks.ParseArenaText(text, name)
	if err != nil {
		return decks.Deck{}, apperr.New(apperr.CodeValidationFailed,
			"decks.parse", err)
	}
	return deck, nil
}

// SaveDeckRevision validates and records one revision of the deck.
func (a *App) SaveDeckRevision(name, text string) viewstate.State {
	deck, err := parseDeck(name, text)
	if err != nil {
		return viewstate.FromError(err)
	}
	service, err := deckService()
	if err != nil {
		return viewstate.FromError(err)
	}
	if _, _, err := service.SaveRevision(a.contextOrBackground(), deck,
		workspaceRuleset()); err != nil {
		return viewstate.FromError(err)
	}
	return viewstate.Success()
}
