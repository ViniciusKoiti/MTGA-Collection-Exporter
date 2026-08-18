package readtools

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/toolreg"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
)

func registry(t *testing.T) *toolreg.Registry {
	t.Helper()
	snapshots := inmem.NewSnapshotStore()
	obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
		time.Unix(1_699_999_000, 0), nil, nil)
	snap, err := collection.NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute),
		[]collection.Entry{
			{Identity: collection.CardIdentity{Printing: "p1", Arena: 101,
				Name: "Lightning Strike", Set: "DMU"}, Quantity: 2},
			{Identity: collection.CardIdentity{Printing: "p2", Arena: 102,
				Name: "Mountain", Set: "UNF"}, Quantity: 40},
			{Unresolved: true, Raw: "???", Quantity: 1},
		}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := snapshots.Save(t.Context(), snap); err != nil {
		t.Fatalf("seed: %v", err)
	}
	catalog := &inmem.Catalog{
		PorArena: map[collection.ArenaID]collection.CardIdentity{
			101: {Printing: "p1", Arena: 101, Name: "Lightning Strike", Set: "DMU"}},
		PorNome: map[string]collection.CardIdentity{
			"Mountain": {Printing: "p2", Arena: 102, Name: "Mountain", Set: "UNF"}},
	}
	ruleset := decks.StandardPadrao(map[string]bool{"lightning strike": true},
		time.Unix(1_700_000_000, 0))
	reg := toolreg.New(0, 0)
	if err := Register(reg, Deps{Snapshots: snapshots, Catalog: catalog, Ruleset: ruleset}); err != nil {
		t.Fatalf("register: %v", err)
	}
	return reg
}

func invoke(t *testing.T, reg *toolreg.Registry, tool, args string) []json.RawMessage {
	t.Helper()
	var raw json.RawMessage
	if args != "" {
		raw = json.RawMessage(args)
	}
	result, err := reg.Invoke(t.Context(), tool, "corr-1", raw, toolreg.Page{})
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	return result.Items
}

func TestReadToolsAnswerThroughTheRegistry(t *testing.T) {
	reg := registry(t)
	if got := string(invoke(t, reg, "sync-status", "")[0]); !strings.Contains(got, `"cards":43`) ||
		!strings.Contains(got, "snap-1") {
		t.Fatalf("sync-status unexpected: %s", got)
	}
	if got := string(invoke(t, reg, "collection-summary", "")[0]); !strings.Contains(got, `"resolved_entries":2`) ||
		!strings.Contains(got, `"unresolved_entries":1`) {
		t.Fatalf("summary unexpected: %s", got)
	}
	hits := invoke(t, reg, "search-cards", `{"q":"light"}`)
	if len(hits) != 1 || !strings.Contains(string(hits[0]), "Lightning Strike") {
		t.Fatalf("search unexpected: %v", hits)
	}
	if got := string(invoke(t, reg, "card-lookup", `{"arena":101}`)[0]); !strings.Contains(got, `"found":true`) {
		t.Fatalf("lookup by arena unexpected: %s", got)
	}
	if got := string(invoke(t, reg, "card-lookup", `{"name":"Ghost"}`)[0]); !strings.Contains(got, `"found":false`) {
		t.Fatalf("lookup miss unexpected: %s", got)
	}
}

func TestCheckDeckReportsGapsAndViolations(t *testing.T) {
	reg := registry(t)
	deckText := "Deck\n4 Lightning Strike (DMU) 137\n30 Mountain (UNF) 240\n"
	items := invoke(t, reg, "check-deck", `{"deck_text":"`+
		strings.ReplaceAll(deckText, "\n", `\n`)+`"}`)
	joined := ""
	for _, item := range items {
		joined += string(item)
	}
	for _, expected := range []string{`"legal":false`, `"missing_total":2`,
		`"gap":"Lightning Strike"`, `"violation":"deck_pequeno"`} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("check-deck missing %s:\n%s", expected, joined)
		}
	}
}
