package collectionsync

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/normalize"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// no adapta uma função tipada sobre Estado ao contrato wf.Node.
type no struct {
	id wf.NodeID
	fn func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error)
}

func (n no) ID() wf.NodeID { return n.id }
func (n no) Execute(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
	estado, outcome, err := n.fn(ctx, estadoDe(st))
	return estado, outcome, err
}

// observa detecta a fonte configurada e captura a observação bruta.
func observa(deps Deps) wf.Node {
	return no{"observa", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		if deps.Source == nil {
			return estado, "sem_fonte", nil
		}
		obs, err := deps.Source.Observe(ctx)
		if err != nil {
			return estado, "", fmt.Errorf("collection-sync: fonte %s falhou: %w",
				deps.Source.Kind(), err)
		}
		estado.Observacao = obs
		return estado, "ok", nil
	}}
}

// normaliza delegates to the single product normalization (task 3.2):
// the graph node owns orchestration only, never the rules.
func normaliza(deps Deps) wf.Node {
	return no{"normaliza", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		result, err := normalize.Observation(ctx, deps.Catalog, estado.Observacao)
		if err != nil {
			return estado, "", err
		}
		estado.Entradas = result.Entries
		estado.Diagnosticos = result.Diagnostics
		estado.Resolvidas = result.Resolved
		estado.NaoResolvidas = result.Unresolved
		return estado, "ok", nil
	}}
}

// valida rejeita coleções vazias antes do commit.
func valida() wf.Node {
	return no{"valida", func(_ context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		if len(estado.Entradas) == 0 {
			return estado, "vazia", nil
		}
		return estado, "ok", nil
	}}
}
