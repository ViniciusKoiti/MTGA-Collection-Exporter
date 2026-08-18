package assistctx

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

var update = flag.Bool("update", false, "rewrite golden files")

func input(t *testing.T) Input {
	t.Helper()
	deck, err := decks.NewDeck("Mono Red", []decks.Entry{
		{Name: "Lightning Strike", Quantity: 4},
		{Name: "Mountain", Quantity: 20},
	}, nil)
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	return Input{
		SyncStatus: "fresh", SnapshotID: "snap-1", TotalCards: 245,
		Question: "how do I finish this deck?", Deck: &deck,
		Ownership: &decks.Ownership{Snapshot: "snap-1", TotalMissing: 2,
			Lines: []decks.OwnershipLine{
				{Name: "Lightning Strike", Required: 4, Owned: 2, Missing: 2},
				{Name: "Mountain", Required: 20, Owned: 20},
			}},
	}
}

// TestContextMatchesRedactionGolden freezes the exact context format:
// any new field shows up as a golden diff and forces a privacy review.
func TestContextMatchesRedactionGolden(t *testing.T) {
	got := Build(input(t))
	golden := filepath.Join("testdata", "context_mono_red.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("testdata: %v", err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("golden missing (run with -update): %v", err)
	}
	if got != string(want) {
		t.Fatalf("context diverged from redaction golden:\n--- got:\n%s--- want:\n%s", got, want)
	}
}

// TestForbiddenContentIsScrubbed laces every field with paths and
// credentials and proves nothing survives into the context.
func TestForbiddenContentIsScrubbed(t *testing.T) {
	in := input(t)
	in.Question = `read C:\Users\me\AppData\player.log and use password=hunter2`
	in.Deck.Main[0].Name = "Bearer abc123 card"
	in.Ownership.Lines[0].Name = "/home/me/.ssh/id_rsa"
	got := Build(in)
	// "/" alone would match the fixed "assistant/v1" header; user data is
	// covered by the concrete path fragments below plus the "\\" check.
	for _, leaked := range []string{"C:", "AppData", "player.log", "hunter2",
		"Bearer", "abc123", ".ssh", "id_rsa", "\\"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("context leaked %q:\n%s", leaked, got)
		}
	}
	if !strings.Contains(got, "[redacted]") {
		t.Fatalf("scrubbing marker missing:\n%s", got)
	}
}

// TestUnrelatedCollectionDataIsAbsentByConstruction: the builder input
// has no field for full collection entries, so only aggregates appear.
func TestUnrelatedCollectionDataIsAbsentByConstruction(t *testing.T) {
	in := input(t)
	in.Deck = nil
	in.Ownership = nil
	got := Build(in)
	if strings.Contains(got, "card|") || strings.Contains(got, "gap|") {
		t.Fatalf("no deck in discussion means no card lines:\n%s", got)
	}
	if !strings.Contains(got, "cards=245") {
		t.Fatalf("aggregate totals are the only collection signal:\n%s", got)
	}
}
