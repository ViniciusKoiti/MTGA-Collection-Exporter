package exportsvc

import (
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// TestStaleApprovalIsDeniedWhenCollectionChanges proves the token binds
// to the exact payload: a snapshot committed after the proposal changes
// the argument hash and the redeem is denied before any effect.
func TestStaleApprovalIsDeniedWhenCollectionChanges(t *testing.T) {
	f := setup(t)
	proposal, err := f.service.ProposeCopy(t.Context(), ports.ExportJSON)
	if err != nil {
		t.Fatalf("proposal: %v", err)
	}
	seed(t, f.snapshots, "snap-2", 9) // collection changed after the proposal
	if _, err := f.service.ExecuteApproved(t.Context(), proposal.TokenID, ports.ExportJSON); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("stale approval must be denied: %v", err)
	}
	if len(f.clipboard.Textos) != 0 {
		t.Fatal("denied execution must not touch the clipboard")
	}
}
