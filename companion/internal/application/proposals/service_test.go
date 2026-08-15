package proposals

import (
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

func fixture(t *testing.T) (*Service, *inmem.SnapshotStore, *inmem.Clock) {
	t.Helper()
	clock := inmem.NewClock(time.Unix(1_700_000_000, 0))
	snapshots := inmem.NewSnapshotStore()
	obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
		clock.Now().Add(-time.Hour), nil, nil)
	snap, err := collection.NewSnapshot("snap-1", obs, clock.Now(),
		[]collection.Entry{{Identity: collection.CardIdentity{Printing: "p1",
			Arena: 101, Name: "Mountain", Set: "UNF"}, Quantity: 20}}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := snapshots.Save(t.Context(), snap); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return New(inmem.NewApprovalService(clock), snapshots, inmem.Exporter{}), snapshots, clock
}

func deckFixture(t *testing.T) decks.Deck {
	t.Helper()
	deck, err := decks.NewDeck("Mono Red", []decks.Entry{{Name: "Mountain", Quantity: 20}}, nil)
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	return deck
}

func TestAllFourProposalsArePreviewOnly(t *testing.T) {
	service, _, _ := fixture(t)
	saveDeck, err := service.ProposeSaveDeck(t.Context(), deckFixture(t))
	if err != nil {
		t.Fatalf("save-deck: %v", err)
	}
	sync, err := service.ProposeSync(t.Context(), collection.SourceJSONImport)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	file, err := service.ProposeFileExport(t.Context(), ports.ExportJSON, "collection.json")
	if err != nil {
		t.Fatalf("file: %v", err)
	}
	clipboard, err := service.ProposeClipboardExport(t.Context(), ports.ExportCSV)
	if err != nil {
		t.Fatalf("clipboard: %v", err)
	}
	for _, proposal := range []Proposal{saveDeck, sync, file, clipboard} {
		if proposal.TokenID == "" || proposal.ArgsHash == "" || proposal.Summary == "" ||
			proposal.ExpiresAt.IsZero() {
			t.Fatalf("proposal incomplete: %+v", proposal)
		}
	}
	if !strings.Contains(file.Summary, "collection.json") ||
		!strings.Contains(clipboard.Summary, "clipboard") {
		t.Fatalf("summaries should name the exact target: %s / %s",
			file.Summary, clipboard.Summary)
	}
}

func TestAuthorizeBindsOperationHashAndSingleUse(t *testing.T) {
	service, _, _ := fixture(t)
	proposal, err := service.ProposeClipboardExport(t.Context(), ports.ExportJSON)
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	if err := service.Authorize(t.Context(), proposal.TokenID, OpExportFile,
		proposal.ArgsHash); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("wrong operation must be denied: %v", err)
	}
	if err := service.Authorize(t.Context(), proposal.TokenID, OpExportClipboard,
		"other-hash"); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("wrong hash must be denied: %v", err)
	}
	if err := service.Authorize(t.Context(), proposal.TokenID, OpExportClipboard,
		proposal.ArgsHash); err != nil {
		t.Fatalf("exact authorization failed: %v", err)
	}
	if err := service.Authorize(t.Context(), proposal.TokenID, OpExportClipboard,
		proposal.ArgsHash); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("token is single-use: %v", err)
	}
}
