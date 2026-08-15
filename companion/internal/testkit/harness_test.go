package testkit

import (
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func executa(t *testing.T, sc Scenario) Result {
	t.Helper()
	h, err := New(sc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	res, err := h.Run(t.Context())
	if err != nil {
		t.Fatalf("run do cenário: %v", err)
	}
	return res
}

func TestCenarioAprovadoConcluiComTransicoesOrdenadas(t *testing.T) {
	res := executa(t, cenarioExport(true))
	if err := res.AssertDesfecho(wf.RunSucceeded, "done"); err != nil {
		t.Fatal(err)
	}
	// valida, exporta pausado (tentativa registrada) e exporta aprovado
	if err := res.AssertTransicoes("valida", "exporta", "exporta"); err != nil {
		t.Fatal(err)
	}
	if err := res.AssertProibido("decks/a.txt", "cartas"); err != nil {
		t.Fatal(err)
	}
}

func TestMesmoCenarioMesmaSeedEvidenciaIdentica(t *testing.T) {
	primeira := executa(t, cenarioExport(true)).Evidence()
	segunda := executa(t, cenarioExport(true)).Evidence()
	if primeira != segunda {
		t.Fatalf("evidência divergiu entre execuções:\n--- 1:\n%s--- 2:\n%s", primeira, segunda)
	}
	if primeira == "" {
		t.Fatal("evidência vazia")
	}
}

func TestCenarioNegadoTerminaComPolicyDenied(t *testing.T) {
	res := executa(t, cenarioExport(false))
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomePolicyDenied); err != nil {
		t.Fatal(err)
	}
}
