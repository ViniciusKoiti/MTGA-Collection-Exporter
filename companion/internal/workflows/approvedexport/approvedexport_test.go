package approvedexport

import (
	"context"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

type exigeAprovacao struct{}

func (exigeAprovacao) Decide(context.Context, wf.Identity, wf.EffectPreview) (wf.PolicyDecision, error) {
	return wf.PolicyRequireApproval, nil
}

// ambiente semeia um snapshot commitado e monta o cenário com outbox
// compartilhado para as asserções de idempotência.
func ambiente(t *testing.T, conceder bool) (testkit.Scenario, *memory.Outbox) {
	t.Helper()
	snapshots := inmem.NewSnapshotStore()
	obs, err := collection.NewObservation(collection.SourceJSONImport, "fixture",
		time.Unix(1_699_999_000, 0), nil, nil)
	if err != nil {
		t.Fatalf("observação: %v", err)
	}
	snap, err := collection.NewSnapshot("snap-1", obs, obs.ObservedAt.Add(time.Minute),
		[]collection.Entry{{Identity: collection.CardIdentity{
			Printing: "p101", Arena: 101, Name: "Carta A", Set: "TST"}, Quantity: 4}}, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if err := snapshots.Save(t.Context(), snap); err != nil {
		t.Fatalf("seed: %v", err)
	}
	outbox := memory.NewOutbox()
	registrar := func(r *wf.Registry, clock *memory.Clock) error {
		return r.Register(Definition(Deps{
			Snapshots: snapshots, Exporter: inmem.Exporter{}, Outbox: outbox, Clock: clock,
		}))
	}
	return testkit.Scenario{
		Nome: "approved-export", Graph: Identity, Registrar: registrar,
		Policy:     exigeAprovacao{},
		Aprovacoes: map[string]testkit.Decisao{"export-write": {Conceder: conceder, Validade: time.Hour}},
		Inicio:     time.Unix(1_700_000_000, 0),
	}, outbox
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

func TestExportAprovadoEnfileiraEfeitoIdempotente(t *testing.T) {
	sc, outbox := ambiente(t, true)
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeDone); err != nil {
		t.Fatal(err)
	}
	if err := res.AssertTransicoes("carrega", "aprova", "aprova", "escreve"); err != nil {
		t.Fatal(err)
	}
	final := res.Run.State.(Estado)
	if final.Hash == "" || final.Bytes == 0 || final.EfeitoID != "export-"+final.Hash {
		t.Fatalf("evidência de conclusão incompleta: %+v", final)
	}
	if res2 := executa(t, sc); res2.Err != nil { // segundo run: mesmo payload
		t.Fatalf("segundo run: %v", res2.Err)
	}
	pendentes, _ := outbox.Pending(t.Context())
	if len(pendentes) != 1 || pendentes[0].ID != final.EfeitoID {
		t.Fatalf("escrita deveria deduplicar pelo hash do payload: %+v", pendentes)
	}
}
