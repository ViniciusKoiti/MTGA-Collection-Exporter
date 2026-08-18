package collectionsync

import (
	"errors"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func TestColecaoVaziaFalhaSemCommit(t *testing.T) {
	sc, snapshots := ambiente(t, fonte(nil, nil))
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunFailed, OutcomeVazia); err != nil {
		t.Fatal(err)
	}
	if _, existe, _ := snapshots.Latest(t.Context()); existe {
		t.Fatal("validação reprovada não pode commitar snapshot")
	}
}

func TestFonteAusenteFalhaDeclarada(t *testing.T) {
	semFonte, _ := ambiente(t, nil)
	semFonte.Registrar = func(r *wf.Registry, clock *memory.Clock) error {
		return r.Register(Definition(Deps{Catalog: &inmem.Catalog{},
			Snapshots: inmem.NewSnapshotStore(), Exporter: inmem.Exporter{}, Clock: clock}))
	}
	if err := executa(t, semFonte).AssertDesfecho(wf.RunFailed, OutcomeSemFonte); err != nil {
		t.Fatal(err)
	}
}

func TestFalhaDeFonteEncerraComoNodeError(t *testing.T) {
	quebrada, _ := ambiente(t, fonte(nil, errors.New("arquivo trancado")))
	res := executa(t, quebrada)
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomeNodeError); err != nil {
		t.Fatal(err)
	}
	if err := res.AssertTransicoes("observa"); err != nil {
		t.Fatal(err)
	}
}
