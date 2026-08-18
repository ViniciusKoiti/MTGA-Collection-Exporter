package devscenario

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
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

// prepara valida o cenário JSON estritamente antes de tocar qualquer
// fixture ou grafo (spec development-harness).
func prepara() wf.Node {
	return no{"prepara", func(_ context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		parsed, err := testkit.ParseScenario([]byte(estado.CenarioJSON))
		if err != nil {
			estado.Relatorio = fmt.Sprintf("cenário rejeitado na validação: %v", err)
			return estado, "invalido", nil
		}
		estado.Parsed = parsed
		return estado, "ok", nil
	}}
}

// executa roda o grafo alvo pelo harness de produção e captura a
// evidência normalizada; nada aqui reimplementa transições.
func executa(deps Deps) wf.Node {
	return no{"executa", func(ctx context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		registrar, existe := deps.Registrars[estado.Parsed.Graph.Kind]
		if !existe {
			estado.Relatorio = fmt.Sprintf("grafo alvo %q fora do catálogo", estado.Parsed.Graph.Kind)
			return estado, "sem_grafo", nil
		}
		cenario, err := estado.Parsed.Materializar(registrar, deps.Policies[estado.Parsed.Graph.Kind])
		if err != nil {
			return estado, "", err
		}
		harness, err := testkit.New(cenario)
		if err != nil {
			return estado, "", err
		}
		resultado, err := harness.Run(ctx)
		if err != nil {
			return estado, "", err
		}
		estado.Desfecho = fmt.Sprintf("%s/%s", resultado.Run.Status, resultado.Run.Outcome)
		estado.Evidencia = resultado.Evidence()
		return estado, "ok", nil
	}}
}
