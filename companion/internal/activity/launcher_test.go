package activity

import (
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func launcher(t *testing.T) *Launcher {
	t.Helper()
	registryGrafos := grafos(t)
	inventario := New()
	if err := inventario.RegisterGraph("sync-collection",
		wf.Identity{Kind: "collection-sync", Version: 1}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := inventario.RegisterPureQuery("check-deck", "aritmética pura"); err != nil {
		t.Fatalf("register pure: %v", err)
	}
	engine := wf.NewEngine(registryGrafos, memory.NewClock(time.Unix(1_700_000_000, 0)),
		memory.NewIDs("run"), memory.NewRunStore())
	l, err := NewLauncher(inventario, registryGrafos, engine)
	if err != nil {
		t.Fatalf("launcher: %v", err)
	}
	return l
}

func TestLauncherExecutaAtividadeDeGrafo(t *testing.T) {
	run, err := launcher(t).Run(t.Context(), "sync-collection", nil)
	if err != nil || run.Status != wf.RunSucceeded {
		t.Fatalf("atividade deveria executar o grafo: %v / %+v", err, run)
	}
}

func TestLauncherRecusaConsultaPuraEDesconhecida(t *testing.T) {
	l := launcher(t)
	if _, err := l.Run(t.Context(), "check-deck", nil); err == nil {
		t.Fatal("consulta pura não inicia grafo")
	}
	if _, err := l.Run(t.Context(), "atividade-fantasma", nil); err == nil {
		t.Fatal("atividade fora do inventário não executa")
	}
}

func TestNewLauncherReprovaInventarioSemGrafoRegistrado(t *testing.T) {
	registryGrafos := grafos(t)
	inventario := New()
	_ = inventario.RegisterGraph("export-approved", wf.Identity{Kind: "approved-export", Version: 1})
	engine := wf.NewEngine(registryGrafos, memory.NewClock(time.Unix(1_700_000_000, 0)),
		memory.NewIDs("run"), memory.NewRunStore())
	if _, err := NewLauncher(inventario, registryGrafos, engine); err == nil {
		t.Fatal("inventário apontando para grafo ausente deveria reprovar o boot")
	}
}
