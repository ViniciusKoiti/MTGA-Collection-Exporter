//go:build windows

package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/collectionsync"
)

// ImportResult is what the UI renders after an import run.
type ImportResult struct {
	RunID      string          `json:"run_id,omitempty"`
	Outcome    string          `json:"outcome,omitempty"`
	Resolved   int             `json:"resolved"`
	Unresolved int             `json:"unresolved"`
	State      viewstate.State `json:"state"`
}

// ImportCollection asks for a legacy JSON export and runs the import.
func (a *App) ImportCollection() ImportResult {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import your MTGA collection export",
		Filters: []runtime.FileFilter{{DisplayName: "MTGA export (*.json)",
			Pattern: "*.json"}}})
	if err != nil {
		return ImportResult{State: viewstate.FromError(err)}
	}
	if path == "" {
		return ImportResult{State: viewstate.Empty()} // user cancelled
	}
	return a.ImportCollectionFrom(path)
}

// ImportCollectionFrom runs the collection-sync graph for one export
// file THROUGH the activity registry — Wails commands never touch the
// workflow engine directly (task 4.6). The path-taking form exists so
// the runtime end-to-end suite can drive the flow without the native
// file dialog.
func (a *App) ImportCollectionFrom(path string) ImportResult {
	launcher, err := a.launcherFor(path)
	if err != nil {
		return ImportResult{State: viewstate.FromError(err)}
	}
	run, err := launcher.Run(context.Background(), "import-collection",
		collectionsync.Estado{})
	if err != nil {
		return ImportResult{State: viewstate.FromError(err)}
	}
	result := ImportResult{RunID: string(run.ID),
		Outcome: string(run.Outcome), State: viewstate.Success()}
	if estado, ok := run.State.(collectionsync.Estado); ok {
		result.Resolved = estado.Resolvidas
		result.Unresolved = estado.NaoResolvidas
	}
	if run.Status != wf.RunSucceeded {
		result.State = viewstate.Stale("import ended as " + string(run.Outcome))
	}
	return result
}
