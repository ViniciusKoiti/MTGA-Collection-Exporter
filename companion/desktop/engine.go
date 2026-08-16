//go:build windows

package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/compatibility"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/legacyjson"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/scryfall"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/collectionsync"
)

// ActivityEvent is the typed progress payload published to the
// frontend for every workflow event (harness task 4.6).
type ActivityEvent struct {
	Run     string            `json:"run"`
	Graph   string            `json:"graph"`
	Step    int               `json:"step"`
	Outcome string            `json:"outcome"`
	At      string            `json:"at"`
	Attrs   map[string]string `json:"attrs,omitempty"`
}

// wailsEvents forwards the allowlisted workflow events to the UI bus.
type wailsEvents struct{ ctx context.Context }

func (s wailsEvents) Emit(_ context.Context, ev wf.Event) error {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "activity:event", ActivityEvent{
			Run: string(ev.Run), Graph: string(ev.Graph.Kind), Step: ev.Step,
			Outcome: string(ev.Outcome), At: ev.At.UTC().Format("15:04:05"),
			Attrs: ev.Attrs})
	}
	return nil
}

// desktopCatalog serves the cached Scryfall catalog when present; an
// absent cache resolves nothing, keeping unresolved entries honest.
func desktopCatalog() ports.Catalog {
	dir, err := scryfallDir()
	if err == nil {
		if catalog, cacheErr := scryfall.OpenCachedOnly(dir); cacheErr == nil {
			return catalog
		}
	}
	return emptyCatalog{}
}

type emptyCatalog struct{}

func (emptyCatalog) ResolvePorArena(context.Context,
	collection.ArenaID) (collection.CardIdentity, bool, error) {
	return collection.CardIdentity{}, false, nil
}

func (emptyCatalog) ResolvePorNome(context.Context,
	string) (collection.CardIdentity, bool, error) {
	return collection.CardIdentity{}, false, nil
}

// launcherFor assembles the real engine for one import: commands only
// ever start graphs through the activity inventory (anti-bypass).
func (a *App) launcherFor(sourcePath string) (*activity.Launcher, error) {
	snapshots, err := collectionsStore()
	if err != nil {
		return nil, err
	}
	graphs := wf.NewRegistry(wf.DefaultLimits())
	if err := graphs.Register(collectionsync.Definition(collectionsync.Deps{
		Source:    &legacyjson.Source{Path: sourcePath, Clock: systemClock{}},
		Catalog:   desktopCatalog(),
		Snapshots: snapshots,
		Exporter:  compatibility.Exporter{},
		Clock:     systemClock{},
	})); err != nil {
		return nil, err
	}
	inventory := activity.New()
	if err := inventory.RegisterGraph("import-collection",
		collectionsync.Identity); err != nil {
		return nil, err
	}
	engine := wf.NewEngine(graphs, systemClock{}, memory.NewIDs("run"),
		store).WithEvents(wailsEvents{ctx: a.ctx})
	return activity.NewLauncher(inventory, graphs, engine)
}
