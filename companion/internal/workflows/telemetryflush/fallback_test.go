package telemetryflush

import (
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func TestSemOptInNadaSaiDoDispositivo(t *testing.T) {
	sc, fila, sink := ambiente(t, false, 0)
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeLocalOnly); err != nil {
		t.Fatal(err)
	}
	if len(sink.Lotes) != 0 || fila.PendentesAgora() != 2 {
		t.Fatal("sem opt-in os eventos permanecem locais e nada é enviado")
	}
}

func TestCentralIndisponivelFazFallbackLocalSemAck(t *testing.T) {
	sc, fila, sink := ambiente(t, true, 99) // falha além do orçamento de retry
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeLocalOnly); err != nil {
		t.Fatal(err)
	}
	if len(sink.Lotes) != 0 || fila.PendentesAgora() != 2 {
		t.Fatal("falha persistente mantém os eventos na fila local")
	}
	if res.Run.State.(Estado).Tentativas != 3 {
		t.Fatalf("retry deveria respeitar o orçamento: %+v", res.Run.State)
	}
}
