package metarecommend

import (
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func TestSemSnapshotFalhaDeclarada(t *testing.T) {
	res := executa(t, cenario(t, nil, false))
	if err := res.AssertDesfecho(wf.RunFailed, OutcomeSemSnapshot); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogoVazioFalhaComoStale(t *testing.T) {
	res := executa(t, cenario(t, nil, true))
	if err := res.AssertDesfecho(wf.RunFailed, OutcomeSemCatalogo); err != nil {
		t.Fatal(err)
	}
}
