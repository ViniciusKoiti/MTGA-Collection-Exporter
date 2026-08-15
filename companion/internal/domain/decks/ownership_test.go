package decks

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

func snapshotCom(t *testing.T, entradas []collection.Entry) collection.Snapshot {
	t.Helper()
	obs, err := collection.NewObservation(collection.SourceJSONImport, "fixture",
		time.Unix(1_700_000_000, 0), nil, nil)
	if err != nil {
		t.Fatalf("observação: %v", err)
	}
	snap, err := collection.NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute), entradas, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

func TestCompareOwnershipSomaImpressoesEIgnoraNaoResolvidas(t *testing.T) {
	deck, err := NewDeck("Teste", []Entry{
		{Name: "Lightning Strike", Quantity: 4},
		{Name: "Mountain", Quantity: 20},
	}, []Entry{{Name: "Lightning Strike", Quantity: 1}}) // side soma com o main
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	snap := snapshotCom(t, []collection.Entry{
		{Identity: collection.CardIdentity{Printing: "p1", Arena: 1, Name: "Lightning Strike", Set: "DMU"}, Quantity: 2},
		{Identity: collection.CardIdentity{Printing: "p2", Arena: 2, Name: "lightning strike", Set: "M21"}, Quantity: 1},
		{Identity: collection.CardIdentity{Printing: "p3", Arena: 3, Name: "Mountain", Set: "UNF"}, Quantity: 40},
		{Unresolved: true, Raw: "Lightning Strike ???", Quantity: 9}, // não conta
	})
	posse := CompareOwnership(deck, snap)
	if posse.Complete || posse.TotalMissing != 2 {
		t.Fatalf("faltantes inesperados: %+v", posse)
	}
	linhas := map[string]OwnershipLine{}
	for _, l := range posse.Lines {
		linhas[l.Name] = l
	}
	strike := linhas["Lightning Strike"]
	if strike.Required != 5 || strike.Owned != 3 || strike.Missing != 2 ||
		strike.WildcardRelevant != 2 {
		t.Fatalf("linha de strike inesperada: %+v", strike)
	}
	montanha := linhas["Mountain"]
	if montanha.Required != 20 || montanha.Owned != 20 || montanha.Missing != 0 {
		t.Fatalf("posse completa deveria saturar no exigido: %+v", montanha)
	}
}

func TestCompareOwnershipDeckCompleto(t *testing.T) {
	deck, _ := NewDeck("Completo", []Entry{{Name: "Mountain", Quantity: 4}}, nil)
	snap := snapshotCom(t, []collection.Entry{
		{Identity: collection.CardIdentity{Printing: "p3", Arena: 3, Name: "Mountain", Set: "UNF"}, Quantity: 4},
	})
	posse := CompareOwnership(deck, snap)
	if !posse.Complete || posse.TotalMissing != 0 || posse.Snapshot != "snap-1" {
		t.Fatalf("deck completo esperado: %+v", posse)
	}
}
