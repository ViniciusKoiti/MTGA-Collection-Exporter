package detailedlogs

import "testing"

// TestParseRecordsOverTheAcceptedFixtures: the parser sees exactly
// the structured records of the sanitized real capture and drops the
// noise.
func TestParseRecordsOverTheAcceptedFixtures(t *testing.T) {
	records := ParseRecords(fixture(t, "current_client_boot.log"))
	methods, payloads := 0, 0
	var sawDeckSummaries, sawInbound bool
	for _, record := range records {
		switch record.Kind {
		case "method":
			methods++
			if record.Method == "DeckGetDeckSummariesV3" {
				sawDeckSummaries = true
			}
			if record.Inbound {
				sawInbound = true
			}
		case "payload":
			payloads++
		}
	}
	if methods != 11 || payloads != 1 {
		t.Fatalf("fixture must parse to 11 methods and 1 payload: %d/%d",
			methods, payloads)
	}
	if !sawDeckSummaries || !sawInbound {
		t.Fatal("method names and directions must be preserved")
	}
}

// TestExtractCollectionMirrorsTheProbeFinding: nothing extracts from
// the current client, and the hypothetical v3 capture yields the full
// map including the final pair after the shape match.
func TestExtractCollectionMirrorsTheProbeFinding(t *testing.T) {
	if cards, ok := ExtractCollection(fixture(t, "current_client_boot.log")); ok {
		t.Fatalf("the current client must extract nothing: %v", cards)
	}
	cards, ok := ExtractCollection(fixture(t, "hypothetical_v3_collection.log"))
	if !ok || len(cards) != 26 {
		t.Fatalf("the v3 capture must extract all 26 cards: %d %v",
			len(cards), ok)
	}
	if cards[100001] != 2 || cards[100026] != 1 {
		t.Fatalf("counts must survive extraction: %v", cards)
	}
}
