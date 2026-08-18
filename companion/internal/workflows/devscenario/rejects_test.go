package devscenario

import (
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func TestCenarioInvalidoEGrafoForaDoCatalogo(t *testing.T) {
	invalido := executaMeta(t, `{"schema":"scenario/v9"}`)
	if err := invalido.AssertDesfecho(wf.RunFailed, OutcomeInvalido); err != nil {
		t.Fatal(err)
	}
	fantasma := executaMeta(t, `{"schema":"scenario/v1","nome":"x",
		"graph":{"kind":"grafo-fantasma","version":1},"inicio_unix":1700000000}`)
	if err := fantasma.AssertDesfecho(wf.RunFailed, OutcomeSemGrafo); err != nil {
		t.Fatal(err)
	}
}
