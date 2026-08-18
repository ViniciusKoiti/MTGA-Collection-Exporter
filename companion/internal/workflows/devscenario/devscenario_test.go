package devscenario

import (
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/collectionsync"
)

// catalogo expõe o grafo de produção collection-sync com fakes semeados.
func catalogo(t *testing.T) Deps {
	t.Helper()
	registrar := func(r *wf.Registry, clock *memory.Clock) error {
		obs, err := collection.NewObservation(collection.SourceJSONImport, "fixture",
			time.Unix(1_699_999_000, 0),
			[]collection.ObservedQuantity{{RawIdentity: "101", Arena: 101, Quantity: 4}}, nil)
		if err != nil {
			return err
		}
		return r.Register(collectionsync.Definition(collectionsync.Deps{
			Source: &inmem.Source{Tipo: collection.SourceJSONImport, Obs: obs},
			Catalog: &inmem.Catalog{PorArena: map[collection.ArenaID]collection.CardIdentity{
				101: {Printing: "p101", Arena: 101, Name: "Carta A", Set: "TST"},
			}},
			Snapshots: inmem.NewSnapshotStore(),
			Exporter:  inmem.Exporter{},
			Clock:     clock,
		}))
	}
	return Deps{Registrars: map[string]Registrar{"collection-sync": registrar}}
}

func executaMeta(t *testing.T, cenarioJSON string) testkit.Result {
	t.Helper()
	registrar := func(r *wf.Registry, _ *memory.Clock) error {
		return r.Register(Definition(catalogo(t)))
	}
	h, err := testkit.New(testkit.Scenario{
		Nome: "meta", Graph: Identity, Registrar: registrar,
		Estado: Estado{CenarioJSON: cenarioJSON},
		Inicio: time.Unix(1_700_000_000, 0),
	})
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	res, err := h.Run(t.Context())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return res
}

func cenarioAlvo(esperado string) string {
	return `{"schema":"scenario/v1","nome":"sync-ok",
		"graph":{"kind":"collection-sync","version":1},
		"inicio_unix":1700000000` + esperado + `}`
}

func TestCenarioAprovadoProduzRelatorioComEvidencia(t *testing.T) {
	res := executaMeta(t, cenarioAlvo(
		`,"esperado":{"status":"succeeded","outcome":"done"}`))
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeDone); err != nil {
		t.Fatal(err)
	}
	relatorio := res.Run.State.(Estado).Relatorio
	for _, trecho := range []string{"veredicto|aprovado", "alvo|collection-sync/v1",
		"desfecho|succeeded/done", "step|01|observa"} {
		if !strings.Contains(relatorio, trecho) {
			t.Fatalf("relatório sem %q:\n%s", trecho, relatorio)
		}
	}
}

func TestAssercaoDivergenteReprova(t *testing.T) {
	res := executaMeta(t, cenarioAlvo(
		`,"esperado":{"status":"failed","outcome":"failed_empty_collection"}`))
	if err := res.AssertDesfecho(wf.RunFailed, OutcomeReprovado); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Run.State.(Estado).Relatorio, "veredicto|reprovado") {
		t.Fatal("relatório deveria registrar a reprovação")
	}
}
