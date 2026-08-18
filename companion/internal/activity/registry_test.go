package activity

import (
	"context"
	"errors"
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

type noOk struct{ id wf.NodeID }

func (n noOk) ID() wf.NodeID { return n.id }
func (n noOk) Execute(context.Context, wf.State) (wf.State, wf.OutcomeCode, error) {
	return nil, "ok", nil
}

func grafos(t *testing.T) *wf.Registry {
	t.Helper()
	r := wf.NewRegistry(wf.Limits{})
	err := r.Register(wf.Definition{
		Identity: wf.Identity{Kind: "collection-sync", Version: 1},
		Initial:  "unico",
		Nodes:    map[wf.NodeID]wf.Node{"unico": noOk{"unico"}},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "unico", Outcome: "ok"}: {Terminal: "done"},
		},
		Recovery: wf.RecoveryResume,
	})
	if err != nil {
		t.Fatalf("registro de grafo: %v", err)
	}
	return r
}

func TestAtividadeDeGrafoValidaContraRuntime(t *testing.T) {
	inventario := New()
	if err := inventario.RegisterGraph("sync-collection", wf.Identity{Kind: "collection-sync", Version: 1}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := inventario.Validate(grafos(t)); err != nil {
		t.Fatalf("validate deveria passar: %v", err)
	}
	entrada, err := inventario.Resolve("sync-collection")
	if err != nil || entrada.Kind != KindGraph {
		t.Fatalf("resolve inesperado: %+v (%v)", entrada, err)
	}
}

func TestRegistrosInvalidosSaoRejeitados(t *testing.T) {
	inventario := New()
	if err := inventario.RegisterPureQuery("busca-carta", ""); !errors.Is(err, ErrActivityInvalida) {
		t.Fatal("consulta pura sem justificativa deveria falhar")
	}
	if err := inventario.RegisterGraph("x", wf.Identity{}); !errors.Is(err, ErrActivityInvalida) {
		t.Fatal("grafo sem identidade deveria falhar")
	}
	_ = inventario.RegisterPureQuery("busca-carta", "consulta somente leitura no snapshot")
	if err := inventario.RegisterPureQuery("busca-carta", "duplicada"); !errors.Is(err, ErrActivityInvalida) {
		t.Fatal("nome duplicado deveria falhar")
	}
	if _, err := inventario.Resolve("inexistente"); !errors.Is(err, ErrActivityInvalida) {
		t.Fatal("atividade não registrada deveria falhar")
	}
}

func TestAtividadeApontandoParaGrafoAusenteFalha(t *testing.T) {
	inventario := New()
	_ = inventario.RegisterGraph("sync-collection", wf.Identity{Kind: "collection-sync", Version: 99})
	if err := inventario.Validate(grafos(t)); !errors.Is(err, ErrActivityInvalida) {
		t.Fatal("grafo ausente no runtime deveria reprovar o inventário")
	}
}
