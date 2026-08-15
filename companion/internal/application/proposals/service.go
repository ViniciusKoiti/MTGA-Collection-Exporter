// Package proposals exposes preview-only proposals for the four
// assistant-requested local effects — deck save, collection sync, file
// export and clipboard export (OpenSpec introduce-agentic-go-companion,
// task 6.5). Proposing NEVER executes anything: it renders an immutable,
// redacted preview and issues an expiring single-use token; execution
// elsewhere must call Authorize with the exact operation and hash first.
package proposals

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Operation names are stable and are part of the token binding.
const (
	OpSaveDeck        = "save-deck"
	OpSyncCollection  = "sync-collection"
	OpExportFile      = "export-file"
	OpExportClipboard = "export-clipboard"
)

// Proposal is the immutable preview presented for approval.
type Proposal struct {
	Operation string
	Summary   string // human-reviewable, redacted
	Target    string
	ArgsHash  string
	TokenID   string
	ExpiresAt time.Time
}

// Service issues proposals and authorizes their execution.
type Service struct {
	approvals ports.ApprovalService
	snapshots ports.SnapshotStore
	exporter  ports.Exporter
}

// New wires the service from the composition root.
func New(approvalSvc ports.ApprovalService, snapshots ports.SnapshotStore,
	exporter ports.Exporter) *Service {
	return &Service{approvals: approvalSvc, snapshots: snapshots, exporter: exporter}
}

// propose hashes the exact effect, requests the token and assembles the
// preview; no port with side effects is touched here.
func (s *Service) propose(ctx context.Context, op, target, summary string,
	payload []byte) (Proposal, error) {
	sum := sha256.Sum256(append([]byte(op+"\x00"+target+"\x00"), payload...))
	hash := fmt.Sprintf("%x", sum[:16])
	token, err := s.approvals.Request(ctx, op, hash)
	if err != nil {
		return Proposal{}, apperr.New(apperr.CodeInternal, "proposals.propose", err)
	}
	return Proposal{Operation: op, Summary: summary, Target: target,
		ArgsHash: hash, TokenID: token.ID, ExpiresAt: token.ExpiresAt}, nil
}

// Authorize redeems the token for the exact operation and hash; every
// executor MUST call it immediately before performing the effect.
func (s *Service) Authorize(ctx context.Context, tokenID, op, argsHash string) error {
	if err := s.approvals.Redeem(ctx, tokenID, op, argsHash); err != nil {
		return apperr.New(apperr.CodeApprovalDenied, "proposals.authorize", err)
	}
	return nil
}
