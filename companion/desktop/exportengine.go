//go:build windows

package main

import (
	"sync"

	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/compatibility"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/approvedexport"
)

// askUserPolicy pauses EVERY effect for an explicit user decision —
// the desktop never writes a file the user did not approve.
type askUserPolicy struct{}

func (askUserPolicy) Decide(context.Context, wf.Identity,
	wf.EffectPreview) (wf.PolicyDecision, error) {
	return wf.PolicyRequireApproval, nil
}

// exportStack keeps the approval-gated export engine alive between
// the start and decide bindings of one session.
type exportStack struct {
	launcher  *activity.Launcher
	engine    *wf.Engine
	approvals *memory.Approvals
	outbox    *memory.Outbox
}

var (
	exportOnce sync.Once
	exportRef  *exportStack
	exportErr  error
)

// exportEngine assembles the approved-export graph over the real
// store; the ONLY way in stays the activity inventory.
func (a *App) exportEngine() (*exportStack, error) {
	exportOnce.Do(func() {
		snapshots, err := collectionsStore()
		if err != nil {
			exportErr = err
			return
		}
		stack := &exportStack{approvals: memory.NewApprovals(),
			outbox: memory.NewOutbox()}
		graphs := wf.NewRegistry(wf.DefaultLimits())
		if err := graphs.Register(approvedexport.Definition(approvedexport.Deps{
			Snapshots: snapshots,
			Exporter:  compatibility.Exporter{},
			Outbox:    stack.outbox,
			Clock:     systemClock{},
		})); err != nil {
			exportErr = err
			return
		}
		inventory := activity.New()
		if err := inventory.RegisterGraph("export-approved",
			approvedexport.Identity); err != nil {
			exportErr = err
			return
		}
		// The run store is in-memory ON PURPOSE: the approval pause and
		// resume happen within one interactive session, and the durable
		// store rehydrates state as untyped JSON, which would void the
		// approved preview hash and re-request approval forever.
		stack.engine = wf.NewEngine(graphs, systemClock{},
			memory.NewIDs("export"), memory.NewRunStore()).
			WithEvents(wailsEvents{ctx: a.ctx})
		stack.engine.WithPolicy(askUserPolicy{}, stack.approvals)
		stack.launcher, err = activity.NewLauncher(inventory, graphs,
			stack.engine)
		if err != nil {
			exportErr = err
			return
		}
		exportRef = stack
	})
	return exportRef, exportErr
}
