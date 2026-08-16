package viewstate

import (
	"errors"
	"strings"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
)

func TestConstructorsRespectTheContract(t *testing.T) {
	states := []State{Empty(), Loading(), Success(),
		Stale("snapshot from yesterday"),
		Partial(apperr.CodeCatalogUnavailable, "3 cards unresolved"),
		FromError(apperr.New(apperr.CodeSourceUnavailable, "collection.sync",
			errors.New("boom")))}
	for _, state := range states {
		if err := state.Valid(); err != nil {
			t.Fatalf("constructor produced an invalid state: %+v (%v)",
				state, err)
		}
	}
}

// TestErrorStatesCarryStableCodesAndNeverTheRawCause: the UI sees the
// code and the presentable operation — a cause holding a filesystem
// path must never leak into Detail.
func TestErrorStatesCarryStableCodesAndNeverTheRawCause(t *testing.T) {
	cause := errors.New(`open C:\Users\someone\AppData\collection.json: denied`)
	state := FromError(apperr.New(apperr.CodeSourceUnavailable,
		"collection.sync", cause))
	if state.Status != StatusError ||
		state.Code != apperr.CodeSourceUnavailable {
		t.Fatalf("typed errors must map to their stable code: %+v", state)
	}
	if strings.Contains(state.Detail, `C:\Users`) {
		t.Fatalf("the raw cause leaked into the view: %q", state.Detail)
	}
	untyped := FromError(errors.New("driver exploded"))
	if untyped.Code != apperr.CodeInternal ||
		strings.Contains(untyped.Detail, "exploded") {
		t.Fatalf("untyped errors must be internal, with nothing leaked: %+v",
			untyped)
	}
	if FromError(nil).Status != StatusSuccess {
		t.Fatal("no error means success")
	}
}

func TestValidRefusesContractViolations(t *testing.T) {
	violations := []State{
		{Status: StatusError},                              // error without code
		{Status: StatusPartial},                            // partial without code
		{Status: StatusSuccess, Code: apperr.CodeConflict}, // code where none belongs
		{Status: Status("spinning")},                       // outside the vocabulary
	}
	for _, state := range violations {
		if state.Valid() == nil {
			t.Fatalf("violation must be refused: %+v", state)
		}
	}
}
