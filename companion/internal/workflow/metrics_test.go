package workflow_test

import (
	"strings"
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func TestMetricasAgregamPorGrafoEOutcomeSemRotulosPrivados(t *testing.T) {
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	metricas := wf.NewMetrics()
	engine.WithEvents(wf.MetricsSink{Metrics: metricas, Next: memory.NewEvents()})

	for range 3 {
		if _, err := engine.Start(t.Context(), identidade(), nil); err != nil {
			t.Fatalf("start: %v", err)
		}
	}
	snapshot := metricas.Snapshot()
	desfechos := snapshot[wf.MetricKey{Kind: "collection-sync", Outcome: "done"}]
	if desfechos.Count != 3 {
		t.Fatalf("esperava 3 desfechos done, veio %+v", snapshot)
	}
	passos := snapshot[wf.MetricKey{Kind: "collection-sync", Outcome: "ok"}]
	if passos.Count != 6 { // 2 steps por run
		t.Fatalf("esperava 6 steps ok agregados, veio %+v", snapshot)
	}
	for chave := range snapshot {
		if strings.Contains(string(chave.Kind), "run-") ||
			strings.Contains(string(chave.Outcome), "run-") {
			t.Fatalf("rótulo de métrica contém identificador de run: %+v", chave)
		}
	}
	if len(snapshot) != 2 { // cardinalidade limitada: kind x outcome
		t.Fatalf("cardinalidade inesperada: %+v", snapshot)
	}
}
