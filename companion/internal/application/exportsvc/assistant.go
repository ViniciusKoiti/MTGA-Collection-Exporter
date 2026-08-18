package exportsvc

import (
	"context"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Proposal is the preview returned to the assistant: the token covers
// exactly this payload hash — a changed collection invalidates it.
type Proposal struct {
	TokenID  string
	ArgsHash string
	Bytes    int
	Snapshot collection.SnapshotID
}

// ProposeCopy is the ASSISTANT path, phase one: render the exact payload,
// request an expiring single-use token bound to it, and audit the request.
// Nothing touches the clipboard here.
func (s *Service) ProposeCopy(ctx context.Context, format ports.ExportFormat) (Proposal, error) {
	payload, hash, snapID, err := s.render(ctx, format)
	if err != nil {
		return Proposal{}, err
	}
	token, err := s.approvals.Request(ctx, toolCopyExport, hash)
	if err != nil {
		return Proposal{}, apperr.New(apperr.CodeInternal, "export.propose", err)
	}
	s.registra(ctx, token.ID, hash, "requested")
	return Proposal{TokenID: token.ID, ArgsHash: hash, Bytes: len(payload), Snapshot: snapID}, nil
}

// ExecuteApproved is phase two: re-render, redeem the token against the
// CURRENT payload hash (stale approvals are denied), then copy and audit.
func (s *Service) ExecuteApproved(ctx context.Context, tokenID string, format ports.ExportFormat) (int, error) {
	payload, hash, _, err := s.render(ctx, format)
	if err != nil {
		return 0, err
	}
	if err := s.approvals.Redeem(ctx, tokenID, toolCopyExport, hash); err != nil {
		s.registra(ctx, tokenID, hash, "denied")
		return 0, apperr.New(apperr.CodeApprovalDenied, "export.execute", err)
	}
	if err := s.clipboard.Write(ctx, string(payload)); err != nil {
		return 0, apperr.New(apperr.CodeInternal, "export.execute", err)
	}
	s.registra(ctx, tokenID, hash, "executed")
	return len(payload), nil
}

// registra appends the redacted audit record; audit failures never abort
// the operation, and records never carry payloads.
func (s *Service) registra(ctx context.Context, correlation, argsHash, outcome string) {
	_ = s.audit.Append(ctx, approvals.AuditRecord{
		Correlation: correlation,
		Tool:        toolCopyExport,
		ArgsHash:    argsHash,
		Outcome:     approvals.AuditOutcome(outcome),
		At:          s.clock.Now(),
	})
}
