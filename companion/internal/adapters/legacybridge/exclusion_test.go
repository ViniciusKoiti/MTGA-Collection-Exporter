package legacybridge

import (
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/policy"
)

// TestBridgeIsExcludedFromAgentTools proves task 3.9's isolation clause:
// the legacy scanner is an explicit user command only — the default-deny
// policy prohibits it for the assistant and it is absent from the
// canonical activity inventory as an agent-invocable graph.
func TestBridgeIsExcludedFromAgentTools(t *testing.T) {
	classifier := policy.NewClassifier(
		[]string{"query-collection", "query-stats", "search-cards", "check-deck"},
		[]string{"export-approved", "sync-collection"},
	)
	for _, name := range []string{"legacy-scan", "legacy-bridge", "memory-scan"} {
		if got := classifier.Classify(name); got != policy.RiskProhibited {
			t.Errorf("%s: assistant policy must prohibit the bridge, got %s", name, got)
		}
	}
	inventory, err := activity.Inventario()
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	for _, name := range []string{"legacy-scan", "legacy-bridge"} {
		if _, err := inventory.Resolve(name); err == nil {
			t.Errorf("%s: the bridge must not be an inventory activity", name)
		}
	}
}
