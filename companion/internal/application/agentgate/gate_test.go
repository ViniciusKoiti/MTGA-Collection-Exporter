package agentgate

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/policy"
)

func fixture(t *testing.T) (*Gate, *inmem.Clock) {
	t.Helper()
	registry := toolreg.New(0, 0)
	echo := func(_ context.Context, _ toolreg.Call) ([]json.RawMessage, error) {
		return []json.RawMessage{json.RawMessage(`{"ok":true}`)}, nil
	}
	schema := json.RawMessage(`{"type":"object"}`)
	for _, name := range []string{"search-cards", "export-approved"} {
		if err := registry.Register(toolreg.Tool{Name: name, Schema: schema, Handler: echo}); err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
	}
	classifier := policy.NewClassifier([]string{"search-cards"}, []string{"export-approved"})
	clock := inmem.NewClock(time.Unix(1_700_000_000, 0))
	return New(registry, classifier, inmem.NewApprovalService(clock),
		&inmem.AuditLog{}, clock), clock
}

func TestReadToolRunsWithoutToken(t *testing.T) {
	gate, _ := fixture(t)
	result, err := gate.Invoke(t.Context(), "search-cards", "c1", nil, "", toolreg.Page{})
	if err != nil || len(result.Items) != 1 {
		t.Fatalf("read tool should run directly: %v", err)
	}
}

func TestEffectToolFullApprovalCycleAndDenials(t *testing.T) {
	gate, clock := fixture(t)
	args := json.RawMessage(`{"format":"json"}`)

	if _, err := gate.Invoke(t.Context(), "export-approved", "c1", args, "", toolreg.Page{}); apperr.CodeOf(err) != apperr.CodeApprovalRequired {
		t.Fatalf("effect without token must require approval: %v", err)
	}
	token, err := gate.Propose(t.Context(), "export-approved", args)
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	otherArgs := json.RawMessage(`{"format":"csv"}`)
	if _, err := gate.Invoke(t.Context(), "export-approved", "c1", otherArgs, token.ID, toolreg.Page{}); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("mismatched arguments must be denied: %v", err)
	}
	if _, err := gate.Invoke(t.Context(), "export-approved", "c1", args, token.ID, toolreg.Page{}); err != nil {
		t.Fatalf("approved execution failed: %v", err)
	}
	if _, err := gate.Invoke(t.Context(), "export-approved", "c1", args, token.ID, toolreg.Page{}); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("token reuse must be denied: %v", err)
	}
	expired, err := gate.Propose(t.Context(), "export-approved", args)
	if err != nil {
		t.Fatalf("second propose: %v", err)
	}
	clock.Advance(time.Hour) // beyond the token validity window
	if _, err := gate.Invoke(t.Context(), "export-approved", "c1", args, expired.ID, toolreg.Page{}); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
		t.Fatalf("expired token must be denied: %v", err)
	}
}

func TestProhibitedAndUnknownAlwaysDenied(t *testing.T) {
	gate, _ := fixture(t)
	for _, name := range []string{"memory-scan", "shell", "unknown-tool"} {
		if _, err := gate.Invoke(t.Context(), name, "c1", nil, "", toolreg.Page{}); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
			t.Errorf("%s: must be denied by default: %v", name, err)
		}
		if _, err := gate.Propose(t.Context(), name, nil); apperr.CodeOf(err) != apperr.CodeApprovalDenied {
			t.Errorf("%s: must not even be proposable: %v", name, err)
		}
	}
}
