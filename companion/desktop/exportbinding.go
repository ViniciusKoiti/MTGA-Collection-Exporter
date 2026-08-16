//go:build windows

package main

import (
	"context"
	"errors"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/viewstate"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/approvedexport"
)

// ExportPreview is what the approval dialog renders: the EXACT
// payload the user is deciding about, nothing else. DecisionKey is
// the effect-gate key (hash of the WHOLE preview) the decision must
// reference — the payload hash alone is display information.
type ExportPreview struct {
	RunID       string          `json:"run_id"`
	Target      string          `json:"target"`
	Hash        string          `json:"hash"`
	Bytes       int             `json:"bytes"`
	DecisionKey string          `json:"decision_key"`
	State       viewstate.State `json:"state"`
}

// StartApprovedExport runs the approved-export graph until it pauses
// on the immutable payload preview awaiting the user's decision.
func (a *App) StartApprovedExport(format string) ExportPreview {
	stack, err := a.exportEngine()
	if err != nil {
		return ExportPreview{State: viewstate.FromError(err)}
	}
	run, err := stack.launcher.Run(context.Background(), "export-approved",
		approvedexport.Estado{Formato: ports.ExportFormat(format)})
	if err != nil && !errors.Is(err, wf.ErrApprovalPending) {
		return ExportPreview{State: viewstate.FromError(err)}
	}
	estado, _ := run.State.(approvedexport.Estado)
	preview := ExportPreview{RunID: string(run.ID), Target: estado.Destino,
		Hash: estado.Hash, Bytes: estado.Bytes,
		DecisionKey: wf.EffectPreview{Effect: "export-write",
			Target: estado.Destino, PayloadHash: estado.Hash}.Hash()}
	if run.Status == wf.RunWaiting {
		preview.State = viewstate.Stale("waiting for your approval")
		return preview
	}
	preview.State = viewstate.FromError(apperr.New(apperr.CodeInternal,
		"export.start", errors.New("unexpected run status")))
	return preview
}

// ExportOutcome is the result of the user's decision.
type ExportOutcome struct {
	Written string          `json:"written,omitempty"`
	State   viewstate.State `json:"state"`
}

// DecideExport applies the decision on the exact approved hash and
// resumes the run; an approval dispatches the outbox, which writes
// the file only while the hash still covers the payload.
func (a *App) DecideExport(runID, hash string, grant bool) ExportOutcome {
	stack, err := a.exportEngine()
	if err != nil {
		return ExportOutcome{State: viewstate.FromError(err)}
	}
	ctx := context.Background()
	deadline := systemClock{}.Now().Add(10 * time.Minute)
	if err := stack.approvals.Decide(ctx, wf.RunID(runID), hash, grant,
		deadline); err != nil {
		return ExportOutcome{State: viewstate.FromError(err)}
	}
	run, err := stack.engine.Resume(ctx, wf.RunID(runID))
	if errors.Is(err, wf.ErrPolicyDenied) {
		return ExportOutcome{State: viewstate.FromError(apperr.New(
			apperr.CodeApprovalDenied, "export.decide", err))}
	}
	if err != nil {
		return ExportOutcome{State: viewstate.FromError(err)}
	}
	if run.Status != wf.RunSucceeded {
		return ExportOutcome{State: viewstate.Stale(
			"export ended as " + string(run.Outcome))}
	}
	written, err := a.dispatchExports(ctx, stack)
	if err != nil {
		return ExportOutcome{State: viewstate.FromError(err)}
	}
	return ExportOutcome{Written: written, State: viewstate.Success()}
}
