package telemetryflush

import (
	"strings"
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// ambiente monta fila com 2 eventos (um com atributo proibido) e o sink.
func ambiente(t *testing.T, optIn bool, falhasSink int) (testkit.Scenario, *inmem.TelemetryQueue, *inmem.TelemetrySink) {
	t.Helper()
	base := time.Unix(1_699_999_000, 0)
	fila := &inmem.TelemetryQueue{Itens: []ports.TelemetryEvent{
		{Seq: 1, Name: "sync_completed", At: base,
			Attrs: map[string]string{"source_kind": "json_import", "collection_path": `C:\Users\u\col.json`}},
		{Seq: 2, Name: "export_completed", At: base,
			Attrs: map[string]string{"format": "json", "gigante": strings.Repeat("x", 200)}},
	}}
	sink := &inmem.TelemetrySink{Falhas: falhasSink}
	registrar := func(r *wf.Registry, _ *memory.Clock) error {
		return r.Register(Definition(Deps{
			Consent: &inmem.Consent{Estado: ports.ConsentState{OptIn: optIn, Purpose: "produto", Version: 1}},
			Queue:   fila, Sink: sink,
			AtributosPermitidos: map[string]bool{"source_kind": true, "format": true, "gigante": true},
			MaxValorBytes:       64, TamanhoLote: 10, MaxTentativas: 3,
		}))
	}
	return testkit.Scenario{Nome: "telemetry-flush", Graph: Identity,
		Registrar: registrar, Inicio: time.Unix(1_700_000_000, 0)}, fila, sink
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

func TestEnvioMinimizadoComAckERetry(t *testing.T) {
	sc, fila, sink := ambiente(t, true, 2) // duas falhas: sucesso na 3a tentativa
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunSucceeded, OutcomeDone); err != nil {
		t.Fatal(err)
	}
	final := res.Run.State.(Estado)
	if final.Tentativas != 3 || final.Enviados != 2 || final.Minimizados != 2 {
		t.Fatalf("retry/minimização inesperados: %+v", final)
	}
	if len(sink.Lotes) != 1 {
		t.Fatalf("um lote deveria chegar ao central: %d", len(sink.Lotes))
	}
	for _, evento := range sink.Lotes[0] {
		for chave, valor := range evento.Attrs {
			if chave == "collection_path" || len(valor) > 64 {
				t.Fatalf("minimização vazou atributo: %s=%s", chave, valor)
			}
		}
	}
	if fila.PendentesAgora() != 0 {
		t.Fatal("ack deveria confirmar os eventos enviados")
	}
}
