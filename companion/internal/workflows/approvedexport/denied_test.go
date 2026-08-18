package approvedexport

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func TestExportNegadoNaoEnfileira(t *testing.T) {
	sc, outbox := ambiente(t, false)
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomePolicyDenied); err != nil {
		t.Fatal(err)
	}
	if pendentes, _ := outbox.Pending(t.Context()); len(pendentes) != 0 {
		t.Fatalf("negação não pode enfileirar efeito: %+v", pendentes)
	}
}

func TestExportSemSnapshotFalhaDeclarado(t *testing.T) {
	registrar := func(r *wf.Registry, clock *memory.Clock) error {
		return r.Register(Definition(Deps{
			Snapshots: inmem.NewSnapshotStore(), // vazio: nada a exportar
			Exporter:  inmem.Exporter{},
			Outbox:    memory.NewOutbox(),
			Clock:     clock,
		}))
	}
	sc := testkit.Scenario{
		Nome: "approved-export-vazio", Graph: Identity, Registrar: registrar,
		Inicio: time.Unix(1_700_000_000, 0),
	}
	if err := executa(t, sc).AssertDesfecho(wf.RunFailed, OutcomeSemSnapshot); err != nil {
		t.Fatal(err)
	}
}
