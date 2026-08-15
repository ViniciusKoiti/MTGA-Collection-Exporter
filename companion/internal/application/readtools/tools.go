// Package readtools registers the read-only assistant tools (OpenSpec
// introduce-agentic-go-companion, task 6.4): sync status, collection
// summary and search, card lookup, and deck validation with ownership
// gaps. Handlers only read through ports; they are consumed by the
// in-app assistant and the development harness via the typed registry —
// no product MCP is shipped.
package readtools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Deps are the read-only ports the tools consume.
type Deps struct {
	Snapshots ports.SnapshotStore
	Catalog   ports.Catalog
	Ruleset   decks.Standard
}

// Register adds the five read tools to the registry.
func Register(reg *toolreg.Registry, deps Deps) error {
	objectSchema := json.RawMessage(`{"type":"object"}`)
	tools := []toolreg.Tool{
		{Name: "sync-status", Description: "current snapshot provenance and totals",
			Schema: objectSchema, Handler: syncStatus(deps)},
		{Name: "collection-summary", Description: "aggregate counts of the snapshot",
			Schema: objectSchema, Handler: summary(deps)},
		{Name: "search-cards",
			Description: "substring search over resolved snapshot entries",
			Schema:      json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}}}`),
			Handler:     search(deps)},
		{Name: "card-lookup",
			Description: "resolve one card by arena id or exact name",
			Schema:      json.RawMessage(`{"type":"object","properties":{"arena":{"type":"integer"},"name":{"type":"string"}}}`),
			Handler:     lookup(deps)},
		{Name: "check-deck",
			Description: "validate Arena deck text: legality and ownership gaps",
			Schema:      json.RawMessage(`{"type":"object","properties":{"deck_text":{"type":"string"}}}`),
			Handler:     checkDeck(deps)},
	}
	for _, tool := range tools {
		if err := reg.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func syncStatus(deps Deps) toolreg.Handler {
	return func(ctx context.Context, _ toolreg.Call) ([]json.RawMessage, error) {
		snap, exists, err := deps.Snapshots.Latest(ctx)
		if err != nil {
			return nil, err
		}
		if !exists {
			return []json.RawMessage{json.RawMessage(`{"status":"not_configured"}`)}, nil
		}
		item, err := json.Marshal(map[string]any{
			"status": "ready", "snapshot": snap.ID, "source": snap.Source,
			"observed_at": snap.ObservedAt, "cards": snap.TotalCartas(),
		})
		return []json.RawMessage{item}, err
	}
}

func summary(deps Deps) toolreg.Handler {
	return func(ctx context.Context, _ toolreg.Call) ([]json.RawMessage, error) {
		snap, exists, err := deps.Snapshots.Latest(ctx)
		if err != nil || !exists {
			return nil, fmt.Errorf("readtools: no snapshot available (%v)", err)
		}
		resolved, unresolved := 0, 0
		for _, entry := range snap.Entries {
			if entry.Unresolved {
				unresolved++
			} else {
				resolved++
			}
		}
		item, err := json.Marshal(map[string]any{"total_cards": snap.TotalCartas(),
			"resolved_entries": resolved, "unresolved_entries": unresolved})
		return []json.RawMessage{item}, err
	}
}
