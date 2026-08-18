package metarecommend

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/decks"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func deckMeta(t *testing.T, nome string, cartas map[string]int) ports.MetaDeck {
	t.Helper()
	var principal []decks.Entry
	for carta, qtd := range cartas {
		principal = append(principal, decks.Entry{Name: carta, Quantity: qtd})
	}
	deck, err := decks.NewDeck(nome, principal, nil)
	if err != nil {
		t.Fatalf("deck %s: %v", nome, err)
	}
	return ports.MetaDeck{Nome: nome, Formato: "standard", Fonte: "prov-1/2026-08", Deck: deck}
}

func cenario(t *testing.T, metas []ports.MetaDeck, comSnapshot bool) testkit.Scenario {
	t.Helper()
	snapshots := inmem.NewSnapshotStore()
	if comSnapshot {
		obs, _ := collection.NewObservation(collection.SourceJSONImport, "fx",
			time.Unix(1_699_999_000, 0), nil, nil)
		snap, err := collection.NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute),
			[]collection.Entry{
				{Identity: collection.CardIdentity{Printing: "p1", Arena: 1, Name: "Mountain", Set: "UNF"}, Quantity: 20},
				{Identity: collection.CardIdentity{Printing: "p2", Arena: 2, Name: "Lightning Strike", Set: "DMU"}, Quantity: 2},
			}, nil)
		if err != nil {
			t.Fatalf("snapshot: %v", err)
		}
		if err := snapshots.Save(t.Context(), snap); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	registrar := func(r *wf.Registry, _ *memory.Clock) error {
		return r.Register(Definition(Deps{
			Snapshots: snapshots, Meta: &inmem.MetaDecks{Itens: metas}, MaxLacunas: 2,
		}))
	}
	return testkit.Scenario{Nome: "meta-recommend", Graph: Identity,
		Registrar: registrar, Inicio: time.Unix(1_700_000_000, 0)}
}

func executa(t *testing.T, sc testkit.Scenario) testkit.Result {
	t.Helper()
	h, err := testkit.New(sc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	res, err := h.Run(t.Context())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return res
}

func TestRankingDeterministicoComLacunasEEvidencia(t *testing.T) {
	metas := []ports.MetaDeck{
		deckMeta(t, "Quase Montável", map[string]int{"Mountain": 20, "Lightning Strike": 4}),
		deckMeta(t, "Montável", map[string]int{"Mountain": 18}),
		deckMeta(t, "Distante", map[string]int{"Carta A": 4, "Carta B": 4, "Carta C": 4}),
	}
	res := executa(t, cenario(t, metas, true))
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeDone); err != nil {
		t.Fatal(err)
	}
	final := res.Run.State.(Estado)
	if final.Avaliados != 3 || final.Snapshot != "snap-1" {
		t.Fatalf("evidência incompleta: %+v", final)
	}
	ordem := []string{"Montável", "Quase Montável", "Distante"}
	for i, esperado := range ordem {
		if final.Recomendacoes[i].Nome != esperado {
			t.Fatalf("ranking fora de ordem: %+v", final.Recomendacoes)
		}
	}
	quase := final.Recomendacoes[1]
	if quase.Montabilidade != 91 || quase.Faltantes != 2 ||
		len(quase.Lacunas) != 1 || quase.Lacunas[0] != "Lightning Strike" {
		t.Fatalf("lacunas/montabilidade inesperadas: %+v", quase)
	}
	distante := final.Recomendacoes[2]
	if len(distante.Lacunas) != 2 { // MaxLacunas limita a citação
		t.Fatalf("limite de lacunas ignorado: %+v", distante)
	}
}
