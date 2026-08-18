package agentgate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/policy"
)

// Invoke executes a tool under the default-deny policy:
//
//   - read tools run directly through the registry;
//   - local-effect tools require redeeming the single-use token issued by
//     Propose for these EXACT arguments — missing, mismatched, expired or
//     reused tokens are denied before any dispatch;
//   - unknown and prohibited tools are always denied.
//
// Every denial and execution is audited with hashes only.
func (g *Gate) Invoke(
	ctx context.Context,
	name, correlation string,
	args json.RawMessage,
	tokenID string,
	page toolreg.Page,
) (toolreg.Result, error) {
	hash := argsHash(name, args)
	switch g.classifier.Classify(name) {
	case policy.RiskRead:
		result, err := g.registry.Invoke(ctx, name, correlation, args, page)
		if err == nil {
			g.record(ctx, correlation, name, hash, approvals.AuditExecuted)
		}
		return result, err
	case policy.RiskLocalEffect:
		if tokenID == "" {
			g.record(ctx, correlation, name, hash, approvals.AuditDenied)
			return toolreg.Result{}, apperr.New(apperr.CodeApprovalRequired, "agent.invoke",
				fmt.Errorf("tool %q requires an approval token", name))
		}
		if err := g.approvals.Redeem(ctx, tokenID, name, hash); err != nil {
			g.record(ctx, tokenID, name, hash, approvals.AuditDenied)
			return toolreg.Result{}, apperr.New(apperr.CodeApprovalDenied, "agent.invoke", err)
		}
		result, err := g.registry.Invoke(ctx, name, correlation, args, page)
		if err == nil {
			g.record(ctx, tokenID, name, hash, approvals.AuditExecuted)
		}
		return result, err
	default:
		g.record(ctx, correlation, name, hash, approvals.AuditDenied)
		return toolreg.Result{}, apperr.New(apperr.CodeApprovalDenied, "agent.invoke",
			fmt.Errorf("tool %q is prohibited by default-deny policy", name))
	}
}
