package activity_test

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/inmem"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/approvedexport"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/collectionsync"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/devscenario"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/metarecommend"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflows/telemetryflush"
)

// TestInventarioCompletoValidaContraOsGrafosReais é o teste de composição
// do produto: TODA atividade de grafo do inventário canônico resolve para
// uma definição real registrada no runtime — o boot reprovaria qualquer
// atividade órfã (tarefas 1.3/2.7 fechando o elo com a seção 4).
func TestInventarioCompletoValidaContraOsGrafosReais(t *testing.T) {
	clock := memory.NewClock(time.Unix(1_700_000_000, 0))
	snapshots := inmem.NewSnapshotStore()
	registry := wf.NewRegistry(wf.Limits{})
	definicoes := []wf.Definition{
		collectionsync.Definition(collectionsync.Deps{
			Source:  &inmem.Source{},
			Catalog: &inmem.Catalog{}, Snapshots: snapshots,
			Exporter: inmem.Exporter{}, Clock: clock,
		}),
		approvedexport.Definition(approvedexport.Deps{
			Snapshots: snapshots, Exporter: inmem.Exporter{},
			Outbox: memory.NewOutbox(), Clock: clock,
		}),
		metarecommend.Definition(metarecommend.Deps{
			Snapshots: snapshots, Meta: &inmem.MetaDecks{},
		}),
		telemetryflush.Definition(telemetryflush.Deps{
			Consent: &inmem.Consent{}, Queue: &inmem.TelemetryQueue{},
			Sink: &inmem.TelemetrySink{}, MaxTentativas: 1, TamanhoLote: 1,
		}),
		devscenario.Definition(devscenario.Deps{}),
	}
	for _, def := range definicoes {
		if err := registry.Register(def); err != nil {
			t.Fatalf("registro de %s: %v", def.Identity, err)
		}
	}
	inventario, err := activity.Inventario()
	if err != nil {
		t.Fatalf("inventário: %v", err)
	}
	engine := wf.NewEngine(registry, clock, memory.NewIDs("run"), memory.NewRunStore())
	if _, err := activity.NewLauncher(inventario, registry, engine); err != nil {
		t.Fatalf("composição completa deveria validar: %v", err)
	}
}
