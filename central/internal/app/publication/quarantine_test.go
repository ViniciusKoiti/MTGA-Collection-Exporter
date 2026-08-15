package publication

import (
	"testing"
)

// TestRunQuarantinesUnsoundCardsWithoutFailing: one unsound card is
// routed to the quarantine sink while the rest publish normally.
func TestRunQuarantinesUnsoundCardsWithoutFailing(t *testing.T) {
	fake := newFakeDeps(5, nil)
	fake.unsoundGrp = 3 // one of the 10 observations normalizes unsound
	man, err := Run(t.Context(), budgets(), "v1", jobs(2), fake.deps())
	if err != nil {
		t.Fatalf("quarantine must not fail the publication: %v", err)
	}
	written := fake.cartasEscritas()
	if len(written) != 9 {
		t.Fatalf("9 sound cards must publish, got %d", len(written))
	}
	for _, card := range written {
		if card.GrpID == 3 {
			t.Fatal("the quarantined card must not be persisted")
		}
	}
	if len(fake.quarantined) != 1 || fake.quarantined[0].GrpID != 3 {
		t.Fatalf("the unsound card must land in quarantine: %+v",
			fake.quarantined)
	}
	if fake.ativacoes != 1 || man.Object.SHA256 == "" {
		t.Fatalf("publication must still activate exactly once: %+v", man)
	}
}
