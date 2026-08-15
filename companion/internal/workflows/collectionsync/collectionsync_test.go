package collectionsync

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// ambiente monta o cenário do harness com os fakes semeados, expondo o
// snapshot store para asserção pós-run.
func ambiente(t *testing.T, source *inmem.Source) (testkit.Scenario, *inmem.SnapshotStore) {
	t.Helper()
	snapshots := inmem.NewSnapshotStore()
	registrar := func(r *wf.Registry, clock *memory.Clock) error {
		return r.Register(Definition(Deps{
			Source: source,
			Catalog: &inmem.Catalog{PorArena: map[collection.ArenaID]collection.CardIdentity{
				101: {Printing: "p101", Arena: 101, Name: "Carta A", Set: "TST"},
			}},
			Snapshots: snapshots,
			Exporter:  inmem.Exporter{},
			Clock:     clock,
		}))
	}
	return testkit.Scenario{
		Nome:      "collection-sync",
		Graph:     Identity,
		Registrar: registrar,
		Inicio:    time.Unix(1_700_000_000, 0),
	}, snapshots
}

func fonte(quantidades []collection.ObservedQuantity, err error) *inmem.Source {
	obs, obsErr := collection.NewObservation(collection.SourceJSONImport, "fixture",
		time.Unix(1_699_999_000, 0), quantidades, nil)
	if obsErr != nil {
		panic(obsErr)
	}
	return &inmem.Source{Tipo: collection.SourceJSONImport, Obs: obs, Err: err}
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

func TestSincronizacaoCompletaCommitaEProjeta(t *testing.T) {
	sc, snapshots := ambiente(t, fonte([]collection.ObservedQuantity{
		{RawIdentity: "101", Arena: 101, Quantity: 4},
		{RawIdentity: "??? misteriosa", Quantity: 1},
		{RawIdentity: "999", Arena: 999, Quantity: 0}, // ignorada com diagnóstico
	}, nil))
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeDone); err != nil {
		t.Fatal(err)
	}
	if err := res.AssertTransicoes("observa", "normaliza", "valida", "persiste", "projeta"); err != nil {
		t.Fatal(err)
	}
	final := res.Run.State.(Estado)
	if final.Resolvidas != 1 || final.NaoResolvidas != 1 || len(final.Projecoes) != 3 {
		t.Fatalf("estado final inesperado: %+v", final)
	}
	snap, existe, err := snapshots.Latest(t.Context())
	if err != nil || !existe || snap.TotalCartas() != 5 {
		t.Fatalf("snapshot commitado inesperado: %+v (%v)", snap, err)
	}
}
