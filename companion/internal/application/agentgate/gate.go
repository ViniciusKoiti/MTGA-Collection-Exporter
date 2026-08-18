// Package agentgate enforces the assistant execution policy around the
// typed tool registry (OpenSpec introduce-agentic-go-companion, task
// 6.2): default-deny classification, exact argument-bound approval
// tokens with expiration and single use, and audited denial paths.
package agentgate

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/policy"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Gate wraps the registry: read tools run directly, local-effect tools
// require a redeemed token, everything else is denied and audited.
type Gate struct {
	registry   *toolreg.Registry
	classifier *policy.Classifier
	approvals  ports.ApprovalService
	audit      ports.AuditLog
	clock      ports.Clock
}

// New wires the gate from the composition root.
func New(registry *toolreg.Registry, classifier *policy.Classifier,
	approvalSvc ports.ApprovalService, audit ports.AuditLog, clock ports.Clock) *Gate {
	return &Gate{registry: registry, classifier: classifier,
		approvals: approvalSvc, audit: audit, clock: clock}
}

// argsHash binds an approval to the exact tool call: tool name plus the
// raw argument bytes.
func argsHash(name string, args json.RawMessage) string {
	sum := sha256.Sum256(append([]byte(name+"\x00"), args...))
	return fmt.Sprintf("%x", sum[:16])
}

// Propose requests an expiring single-use token for a local-effect tool
// call with these exact arguments; read tools need no proposal and
// prohibited tools cannot even be proposed.
func (g *Gate) Propose(ctx context.Context, name string, args json.RawMessage) (approvals.Token, error) {
	if g.classifier.Classify(name) != policy.RiskLocalEffect {
		g.record(ctx, "propose", name, argsHash(name, args), approvals.AuditDenied)
		return approvals.Token{}, apperr.New(apperr.CodeApprovalDenied, "agent.propose",
			fmt.Errorf("tool %q is not an approvable local effect", name))
	}
	token, err := g.approvals.Request(ctx, name, argsHash(name, args))
	if err != nil {
		return approvals.Token{}, apperr.New(apperr.CodeInternal, "agent.propose", err)
	}
	g.record(ctx, token.ID, name, token.ArgsHash, approvals.AuditRequested)
	return token, nil
}

// record appends the redacted audit record; audit failure never aborts.
func (g *Gate) record(ctx context.Context, correlation, tool, hash string,
	outcome approvals.AuditOutcome) {
	_ = g.audit.Append(ctx, approvals.AuditRecord{Correlation: correlation,
		Tool: tool, ArgsHash: hash, Outcome: outcome, At: g.clock.Now()})
}
