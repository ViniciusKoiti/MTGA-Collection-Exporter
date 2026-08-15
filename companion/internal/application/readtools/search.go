package readtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func search(deps Deps) toolreg.Handler {
	return func(ctx context.Context, call toolreg.Call) ([]json.RawMessage, error) {
		var args struct {
			Q string `json:"q"`
		}
		if len(call.Args) > 0 {
			if err := json.Unmarshal(call.Args, &args); err != nil {
				return nil, err
			}
		}
		snap, exists, err := deps.Snapshots.Latest(ctx)
		if err != nil || !exists {
			return nil, fmt.Errorf("readtools: no snapshot available (%v)", err)
		}
		needle := strings.ToLower(args.Q)
		var items []json.RawMessage
		for _, entry := range snap.Entries {
			if entry.Unresolved || !strings.Contains(strings.ToLower(entry.Identity.Name), needle) {
				continue
			}
			item, err := json.Marshal(map[string]any{
				"name": entry.Identity.Name, "set": entry.Identity.Set,
				"arena": entry.Identity.Arena, "quantity": entry.Quantity,
			})
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	}
}

func lookup(deps Deps) toolreg.Handler {
	return func(ctx context.Context, call toolreg.Call) ([]json.RawMessage, error) {
		var args struct {
			Arena collection.ArenaID `json:"arena"`
			Name  string             `json:"name"`
		}
		if len(call.Args) > 0 {
			if err := json.Unmarshal(call.Args, &args); err != nil {
				return nil, err
			}
		}
		var identity collection.CardIdentity
		var found bool
		var err error
		switch {
		case args.Arena != 0:
			identity, found, err = deps.Catalog.ResolvePorArena(ctx, args.Arena)
		case args.Name != "":
			identity, found, err = deps.Catalog.ResolvePorNome(ctx, args.Name)
		default:
			return nil, fmt.Errorf("readtools: card-lookup requires arena or name")
		}
		if err != nil {
			return nil, err
		}
		if !found {
			return []json.RawMessage{json.RawMessage(`{"found":false}`)}, nil
		}
		item, err := json.Marshal(map[string]any{"found": true, "name": identity.Name,
			"set": identity.Set, "arena": identity.Arena, "printing": identity.Printing})
		return []json.RawMessage{item}, err
	}
}
