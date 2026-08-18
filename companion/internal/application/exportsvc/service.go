// Package exportsvc implements user copy/export commands and the
// approval-token path for assistant-requested operations (OpenSpec
// introduce-agentic-go-companion, task 5.7). Direct user commands run
// immediately; assistant requests are proposal-only until a token bound
// to the exact payload is redeemed.
package exportsvc

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// toolCopyExport is the stable tool name bound to approval tokens.
const toolCopyExport = "copy-export"

// Service wires the ports used by copy/export flows.
type Service struct {
	snapshots ports.SnapshotStore
	exporter  ports.Exporter
	clipboard ports.Clipboard
	approvals ports.ApprovalService
	audit     ports.AuditLog
	clock     ports.Clock
}

// New builds the service from the composition root.
func New(
	snapshots ports.SnapshotStore,
	exporter ports.Exporter,
	clipboard ports.Clipboard,
	approvals ports.ApprovalService,
	audit ports.AuditLog,
	clock ports.Clock,
) *Service {
	return &Service{snapshots: snapshots, exporter: exporter, clipboard: clipboard,
		approvals: approvals, audit: audit, clock: clock}
}

// render projects the current snapshot and returns the payload plus the
// exact argument hash that approval tokens bind to.
func (s *Service) render(
	ctx context.Context,
	format ports.ExportFormat,
) ([]byte, string, collection.SnapshotID, error) {
	snap, exists, err := s.snapshots.Latest(ctx)
	if err != nil {
		return nil, "", "", apperr.New(apperr.CodeInternal, "export.render", err)
	}
	if !exists {
		return nil, "", "", apperr.New(apperr.CodeSnapshotNotFound, "export.render", nil)
	}
	payload, err := s.exporter.Export(ctx, snap, format)
	if err != nil {
		return nil, "", "", apperr.New(apperr.CodeInternal, "export.render", err)
	}
	sum := sha256.Sum256(append([]byte(string(format)+"\x00"), payload...))
	return payload, fmt.Sprintf("%x", sum[:16]), snap.ID, nil
}

// CopyToClipboard is the DIRECT user command: user intent is the
// approval, so it executes immediately and is audited as executed.
func (s *Service) CopyToClipboard(ctx context.Context, format ports.ExportFormat) (int, error) {
	payload, hash, _, err := s.render(ctx, format)
	if err != nil {
		return 0, err
	}
	if err := s.clipboard.Write(ctx, string(payload)); err != nil {
		return 0, apperr.New(apperr.CodeInternal, "export.copy", err)
	}
	s.registra(ctx, "user", hash, "executed")
	return len(payload), nil
}
