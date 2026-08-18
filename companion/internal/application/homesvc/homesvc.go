// Package homesvc composes the Home view model (OpenSpec
// introduce-agentic-go-companion, task 4.3): MTGA presence, source
// health, last sync, freshness, snapshot totals, delta and the
// in-context recovery actions — all through ports, no adapter here.
package homesvc

import (
	"context"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// StaleAfter is the freshness budget: an older snapshot renders stale.
const StaleAfter = 24 * time.Hour

// SnapshotReader is the minimal read surface Home needs.
type SnapshotReader interface {
	Latest(ctx context.Context) (collection.Snapshot, bool, error)
}

// HistoryReader is optional: with it, Home shows the delta against
// the previous snapshot; without it, the delta is honestly unknown.
type HistoryReader interface {
	Previous(ctx context.Context) (collection.Snapshot, bool, error)
}

// Deps wires the ports; the desktop injects the real adapters.
type Deps struct {
	Presence func(ctx context.Context) bool
	Source   func(ctx context.Context) (collection.SourceKind, bool)
	Reader   SnapshotReader
	Now      func() time.Time
}

// Action is one in-context recovery action Home offers.
type Action struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Model is everything the Home view renders.
type Model struct {
	MTGARunning   bool                  `json:"mtga_running"`
	Source        collection.SourceKind `json:"source"`
	SourceHealthy bool                  `json:"source_healthy"`
	LastSync      string                `json:"last_sync,omitempty"`
	Totals        int                   `json:"totals"`
	Delta         *int                  `json:"delta,omitempty"`
	State         viewstate.State       `json:"state"`
	Actions       []Action              `json:"actions"`
}

// Compose builds the Home model; every failure path still renders —
// with its state and a recovery action, never a blank screen.
func Compose(ctx context.Context, deps Deps) Model {
	model := Model{MTGARunning: deps.Presence(ctx)}
	model.Source, model.SourceHealthy = deps.Source(ctx)
	latest, ok, err := deps.Reader.Latest(ctx)
	if err != nil {
		model.State = viewstate.FromError(err)
		model.Actions = []Action{{ID: "retry_sync", Label: "Retry sync"}}
		return model
	}
	if !ok {
		model.State = viewstate.Empty()
		model.Actions = []Action{{ID: "import_collection",
			Label: "Import your collection"}}
		return model
	}
	model.LastSync = latest.ImportedAt.UTC().Format(time.RFC3339)
	for _, entry := range latest.Entries {
		model.Totals += entry.Quantity
	}
	if history, hasHistory := deps.Reader.(HistoryReader); hasHistory {
		if previous, okPrev, errPrev := history.Previous(ctx); errPrev == nil && okPrev {
			delta := model.Totals
			for _, entry := range previous.Entries {
				delta -= entry.Quantity
			}
			model.Delta = &delta
		}
	}
	if deps.Now().Sub(latest.ImportedAt) > StaleAfter {
		model.State = viewstate.Stale("last sync is older than 24h")
		model.Actions = []Action{{ID: "resync", Label: "Sync now"}}
		return model
	}
	model.State = viewstate.Success()
	return model
}
