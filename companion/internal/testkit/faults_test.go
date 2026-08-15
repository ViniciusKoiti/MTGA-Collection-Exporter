package testkit

import (
	"context"
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// cenarioComFalha monta coleta -> valida -> done aplicando o decorador no
// nó de coleta.
func cenarioComFalha(decora func(wf.Node, *memory.Clock) wf.Node) Scenario {
	registrar := func(r *wf.Registry, clock *memory.Clock) error {
		coleta := noFn{"coleta", func(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
			return st, "ok", nil
		}}
		valida := noFn{"valida", func(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
			return st, "ok", nil
		}}
		return r.Register(wf.Definition{
			Identity: wf.Identity{Kind: "collection-sync", Version: 1},
			Initial:  "coleta",
			Nodes:    map[wf.NodeID]wf.Node{"coleta": decora(coleta, clock), "valida": valida},
			Transitions: map[wf.TransitionKey]wf.Target{
				{From: "coleta", Outcome: "ok"}: {Next: "valida"},
				{From: "valida", Outcome: "ok"}: {Terminal: "done"},
			},
			Recovery: wf.RecoveryResume,
		})
	}
	return Scenario{
		Nome:      "collection-sync-com-falha",
		Graph:     wf.Identity{Kind: "collection-sync", Version: 1},
		Registrar: registrar,
		Inicio:    time.Unix(1_700_000_000, 0),
	}
}

func TestLatenciaInjetadaEstouraDeadline(t *testing.T) {
	res := executa(t, cenarioComFalha(func(n wf.Node, c *memory.Clock) wf.Node {
		return ComLatencia(n, c, 45*time.Second) // deadline default: 30s
	}))
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomeDeadlineExceeded); err != nil {
		t.Fatal(err)
	}
}

func TestTimeoutDaFronteiraFalhaComoNodeError(t *testing.T) {
	res := executa(t, cenarioComFalha(func(n wf.Node, _ *memory.Clock) wf.Node {
		return ComTimeout(n)
	}))
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomeNodeError); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(res.Err, context.DeadlineExceeded) {
		t.Fatalf("erro deveria preservar a causa: %v", res.Err)
	}
}

func TestRespostaMalformadaViraOutcomeUnmapped(t *testing.T) {
	res := executa(t, cenarioComFalha(func(n wf.Node, _ *memory.Clock) wf.Node {
		return ComRespostaMalformada(n)
	}))
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomeUnmapped); err != nil {
		t.Fatal(err)
	}
}

func TestDesconexaoDaFronteiraFalhaComCausaPreservada(t *testing.T) {
	res := executa(t, cenarioComFalha(func(n wf.Node, _ *memory.Clock) wf.Node {
		return ComDesconexao(n, 1)
	}))
	if err := res.AssertDesfecho(wf.RunFailed, wf.OutcomeNodeError); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(res.Err, ErrDesconectado) {
		t.Fatalf("erro deveria preservar a desconexão: %v", res.Err)
	}
}
