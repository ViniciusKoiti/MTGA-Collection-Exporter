//go:build windows

package main

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/collectionsvc"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/homesvc"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
)

// CollectionPage is what the Collection view renders for one query.
type CollectionPage struct {
	Rows  []collectionsvc.Row `json:"rows"`
	State viewstate.State     `json:"state"`
}

// Collection filters and sorts the latest snapshot.
func (a *App) Collection(query collectionsvc.Query) CollectionPage {
	reader, err := collections()
	if err != nil {
		return CollectionPage{State: viewstate.FromError(err)}
	}
	latest, ok, err := reader.Latest(context.Background())
	if err != nil {
		return CollectionPage{State: viewstate.FromError(err)}
	}
	if !ok {
		return CollectionPage{State: viewstate.Empty()}
	}
	return CollectionPage{Rows: collectionsvc.Page(latest, query),
		State: viewstate.Success()}
}

// CollectionDiff is the snapshot comparison, or its empty state when
// there is nothing to compare against.
type CollectionDiff struct {
	Diff  collectionsvc.Diff `json:"diff"`
	State viewstate.State    `json:"state"`
}

// CompareSnapshots diffs the latest snapshot against the previous one.
func (a *App) CompareSnapshots() CollectionDiff {
	reader, err := collections()
	if err != nil {
		return CollectionDiff{State: viewstate.FromError(err)}
	}
	ctx := context.Background()
	latest, ok, err := reader.Latest(ctx)
	if err != nil {
		return CollectionDiff{State: viewstate.FromError(err)}
	}
	if !ok {
		return CollectionDiff{State: viewstate.Empty()}
	}
	history, hasHistory := reader.(homesvc.HistoryReader)
	if !hasHistory {
		return CollectionDiff{State: viewstate.Empty()}
	}
	previous, okPrev, err := history.Previous(ctx)
	if err != nil {
		return CollectionDiff{State: viewstate.FromError(err)}
	}
	if !okPrev {
		return CollectionDiff{State: viewstate.Empty()}
	}
	return CollectionDiff{Diff: collectionsvc.Compare(previous, latest),
		State: viewstate.Success()}
}
