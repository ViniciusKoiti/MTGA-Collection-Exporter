package detailedlogs

import (
	"os"
	"regexp"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return raw
}

// TestCurrentClientCarriesNoCollectionSoImportStaysTheSource is the
// documented finding of task 1.2: the sanitized capture of the real
// 2026-08 client has decks, inventory and rank — but no collection
// map — so the decision (task 1.3) falls back to explicit import.
func TestCurrentClientCarriesNoCollectionSoImportStaysTheSource(t *testing.T) {
	caps := Probe(fixture(t, "current_client_boot.log"))
	if caps.CollectionPayload || caps.PayloadVersion != "" {
		t.Fatalf("the current client must probe empty: %+v", caps)
	}
	if src := DecideSource(caps, RecognizedPayloadVersions()); src != collection.SourceJSONImport {
		t.Fatalf("without a collection payload the source must be explicit import: %s", src)
	}
}

// TestRecognizedFuturePayloadEnablesLogsButUnknownOnesNever: the
// enable path exists and works, yet an unrecognized version — even
// carrying a full collection — still falls back.
func TestRecognizedFuturePayloadEnablesLogsButUnknownOnesNever(t *testing.T) {
	caps := Probe(fixture(t, "hypothetical_v3_collection.log"))
	if !caps.CollectionPayload || caps.PayloadVersion != "v3" {
		t.Fatalf("the hypothetical capture must probe full: %+v", caps)
	}
	if src := DecideSource(caps, map[string]bool{"v3": true}); src != collection.SourceDetailedLogs {
		t.Fatalf("a recognized version with the payload must enable logs: %s", src)
	}
	if src := DecideSource(caps, RecognizedPayloadVersions()); src != collection.SourceJSONImport {
		t.Fatalf("the shipping allowlist must not enable unknown versions: %s", src)
	}
}

// TestFixturesAreSanitized: every GUID in the fixtures is the
// synthetic zero-prefixed form — nothing from a real account ships.
func TestFixturesAreSanitized(t *testing.T) {
	guid := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	for _, name := range []string{"current_client_boot.log",
		"hypothetical_v3_collection.log"} {
		for _, id := range guid.FindAllString(string(fixture(t, name)), -1) {
			if id[:19] != "00000000-0000-4000-" {
				t.Fatalf("%s carries a non-synthetic GUID: %s", name, id)
			}
		}
	}
}
