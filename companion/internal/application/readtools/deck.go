package readtools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

// checkDeck parses Arena deck text and returns one summary item followed
// by one item per ownership gap and per rule violation — the registry
// paginates the listing.
func checkDeck(deps Deps) toolreg.Handler {
	return func(ctx context.Context, call toolreg.Call) ([]json.RawMessage, error) {
		var args struct {
			DeckText string `json:"deck_text"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil || args.DeckText == "" {
			return nil, fmt.Errorf("readtools: check-deck requires deck_text")
		}
		deck, err := decks.ParseArenaText(args.DeckText, "assistant-deck")
		if err != nil {
			return nil, err
		}
		snap, exists, err := deps.Snapshots.Latest(ctx)
		if err != nil || !exists {
			return nil, fmt.Errorf("readtools: no snapshot available (%v)", err)
		}
		ownership := decks.CompareOwnership(deck, snap)
		verdict := deps.Ruleset.Validate(deck, deps.Ruleset.CatalogoAtualizado)
		summary, err := json.Marshal(map[string]any{
			"deck": deck.Name, "snapshot": ownership.Snapshot,
			"ruleset": verdict.Ruleset.String(), "legal": verdict.Legal,
			"complete": ownership.Complete, "missing_total": ownership.TotalMissing,
		})
		if err != nil {
			return nil, err
		}
		items := []json.RawMessage{summary}
		for _, line := range ownership.Lines {
			if line.Missing == 0 {
				continue
			}
			gap, err := json.Marshal(map[string]any{"gap": line.Name,
				"required": line.Required, "owned": line.Owned, "missing": line.Missing})
			if err != nil {
				return nil, err
			}
			items = append(items, gap)
		}
		for _, problem := range verdict.Problemas {
			violation, err := json.Marshal(map[string]any{"violation": problem.Code,
				"card": problem.Carta, "detail": problem.Detail})
			if err != nil {
				return nil, err
			}
			items = append(items, violation)
		}
		return items, nil
	}
}
