package toolreg

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/policy"
)

// dangerousCapabilities is the explicit denylist required by task 6.6:
// each one must be impossible to register AND denied on invocation AND
// classified as prohibited by the default-deny policy.
var dangerousCapabilities = []string{
	"memory-scan", "read-process-memory",
	"input-synthesis", "send-input",
	"gameplay-automation", "play-card",
	"purchase-pack", "buy-gems",
	"account-manage", "credential-read", "password-export",
	"shell", "exec-command",
	"filesystem-browse", "file-write-anywhere",
}

func TestDangerousToolsCannotBeRegistered(t *testing.T) {
	registry := New(0, 0)
	noop := func(context.Context, Call) ([]json.RawMessage, error) { return nil, nil }
	schema := json.RawMessage(`{"type":"object"}`)
	for _, name := range dangerousCapabilities {
		err := registry.Register(Tool{Name: name, Schema: schema, Handler: noop})
		if err == nil {
			t.Errorf("%s: registration must be impossible", name)
		}
	}
}

func TestDangerousToolsAreDeniedOnInvocation(t *testing.T) {
	registry := New(0, 0)
	for _, name := range dangerousCapabilities {
		if _, err := registry.Invoke(t.Context(), name, "corr", nil, Page{}); err == nil {
			t.Errorf("%s: invocation must be denied", name)
		}
	}
}

func TestPolicyClassifiesDangerousToolsAsProhibited(t *testing.T) {
	classifier := policy.NewClassifier(
		[]string{"query-collection", "query-stats", "search-cards", "check-deck"},
		[]string{"export-approved", "sync-collection"},
	)
	for _, name := range dangerousCapabilities {
		if got := classifier.Classify(name); got != policy.RiskProhibited {
			t.Errorf("%s: expected prohibited, got %s", name, got)
		}
	}
}
