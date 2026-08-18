package proposals

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// ProposeSaveDeck previews saving a deck revision: name and card counts
// only — the deck list itself stays out of the preview text.
func (s *Service) ProposeSaveDeck(ctx context.Context, deck decks.Deck) (Proposal, error) {
	summary := fmt.Sprintf("save deck %q (%d main, %d sideboard) as a new revision",
		deck.Name, len(deck.Main), len(deck.Sideboard))
	payload := []byte(deck.ArenaText())
	return s.propose(ctx, OpSaveDeck, deck.Name, summary, payload)
}

// ProposeSync previews a collection sync from the given source kind.
func (s *Service) ProposeSync(ctx context.Context, source collection.SourceKind) (Proposal, error) {
	summary := fmt.Sprintf("synchronize the collection from source %s", source)
	return s.propose(ctx, OpSyncCollection, string(source), summary, nil)
}

// renderExport projects the current snapshot for export previews.
func (s *Service) renderExport(ctx context.Context, format ports.ExportFormat) ([]byte, collection.SnapshotID, error) {
	snap, exists, err := s.snapshots.Latest(ctx)
	if err != nil {
		return nil, "", apperr.New(apperr.CodeInternal, "proposals.render", err)
	}
	if !exists {
		return nil, "", apperr.New(apperr.CodeSnapshotNotFound, "proposals.render", nil)
	}
	payload, err := s.exporter.Export(ctx, snap, format)
	if err != nil {
		return nil, "", apperr.New(apperr.CodeInternal, "proposals.render", err)
	}
	return payload, snap.ID, nil
}

// ProposeFileExport previews writing the compatibility projection to the
// named file; the token binds to the exact bytes that would be written.
func (s *Service) ProposeFileExport(ctx context.Context, format ports.ExportFormat,
	filename string) (Proposal, error) {
	payload, snapID, err := s.renderExport(ctx, format)
	if err != nil {
		return Proposal{}, err
	}
	summary := fmt.Sprintf("write %d bytes (%s of snapshot %s) to %q",
		len(payload), format, snapID, filename)
	return s.propose(ctx, OpExportFile, filename, summary, payload)
}

// ProposeClipboardExport previews copying the projection to the clipboard.
func (s *Service) ProposeClipboardExport(ctx context.Context,
	format ports.ExportFormat) (Proposal, error) {
	payload, snapID, err := s.renderExport(ctx, format)
	if err != nil {
		return Proposal{}, err
	}
	summary := fmt.Sprintf("copy %d bytes (%s of snapshot %s) to the clipboard",
		len(payload), format, snapID)
	return s.propose(ctx, OpExportClipboard, "clipboard", summary, payload)
}
