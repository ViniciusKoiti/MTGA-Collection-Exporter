// Package detailedlogs decides whether the MTGA Detailed Logs can
// serve as a collection source (OpenSpec introduce-agentic-go-
// companion, tasks 1.2/1.3).
//
// Finding, documented from real captures of the 2026-08 client
// (three sessions, Detailed Logs enabled, collection screen opened):
// the log carries inventory currencies, deck summaries, rank, quests
// and match traffic, but NO complete grpId->count collection payload
// and no PlayerInventory.GetPlayerCards* method at all. A complete
// collection payload therefore CANNOT be detected reliably, and the
// log source stays disabled — explicit import (or the legacy bridge)
// remains the way in.
package detailedlogs

import (
	"regexp"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Capabilities is what probing one detailed log establishes.
type Capabilities struct {
	CollectionPayload bool   // a complete grpId->count map was found
	PayloadVersion    string // recognized payload marker, "" if none
}

var (
	// collectionShape matches a run of at least twenty grpId->count
	// pairs — the smallest thing we would call a collection map.
	collectionShape = regexp.MustCompile(`("\d{5,6}":\d{1,2},){20,}`)
	versionMarker   = regexp.MustCompile(`GetPlayerCardsV(\d)`)
)

// Probe scans a detailed-log excerpt for a complete collection map
// and the payload version marker that shaped it.
func Probe(log []byte) Capabilities {
	caps := Capabilities{CollectionPayload: collectionShape.Match(log)}
	if m := versionMarker.FindSubmatch(log); m != nil {
		caps.PayloadVersion = "v" + string(m[1])
	}
	return caps
}

// RecognizedPayloadVersions is the shipping allowlist. It is EMPTY on
// purpose: no current client emits the collection payload (verified
// against the real captures in testdata), so the log source cannot
// enable itself by accident. A future client that brings the payload
// back gets added here only after its fixture lands.
func RecognizedPayloadVersions() map[string]bool {
	return map[string]bool{}
}

// DecideSource enables the log source ONLY for a recognized payload
// version that actually carries the collection; everything else
// falls back to the explicit JSON import (task 1.3).
func DecideSource(caps Capabilities,
	recognized map[string]bool) collection.SourceKind {
	if caps.CollectionPayload && recognized[caps.PayloadVersion] {
		return collection.SourceDetailedLogs
	}
	return collection.SourceJSONImport
}
