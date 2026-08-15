package testkit

import (
	"context"
	"errors"
	"fmt"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// Decoradores de falha da tarefa 5.4: envolvem nós e stores DE PRODUÇÃO
// para injetar latência, timeout, resposta malformada, desconexão e crash
// de checkpoint de forma determinística.

// ErrDesconectado simula queda de conexão de um adaptador de fronteira.
var ErrDesconectado = errors.New("testkit: conexão perdida com a fronteira")

type noDecorado struct {
	base wf.Node
	fn   func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error)
}

func (n noDecorado) ID() wf.NodeID { return n.base.ID() }
func (n noDecorado) Execute(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
	return n.fn(ctx, st)
}

// ComLatencia avança o relógio controlado antes de delegar, simulando um
// nó lento sem dormir de verdade.
func ComLatencia(base wf.Node, clock *memory.Clock, d time.Duration) wf.Node {
	return noDecorado{base, func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		clock.Advance(d)
		return base.Execute(ctx, st)
	}}
}

// ComTimeout faz o nó devolver estouro de prazo da fronteira.
func ComTimeout(base wf.Node) wf.Node {
	return noDecorado{base, func(context.Context, wf.State) (wf.State, wf.OutcomeCode, error) {
		return nil, "", fmt.Errorf("testkit: fronteira excedeu o prazo: %w",
			context.DeadlineExceeded)
	}}
}

// ComRespostaMalformada troca o outcome por um código sem transição
// compilada, simulando resposta fora do contrato.
func ComRespostaMalformada(base wf.Node) wf.Node {
	return noDecorado{base, func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		estado, _, err := base.Execute(ctx, st)
		if err != nil {
			return estado, "", err
		}
		return estado, "resposta_malformada_fora_do_contrato", nil
	}}
}

// ComDesconexao falha as primeiras `falhas` execuções com erro de conexão.
func ComDesconexao(base wf.Node, falhas int) wf.Node {
	restantes := falhas
	return noDecorado{base, func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		if restantes > 0 {
			restantes--
			return nil, "", ErrDesconectado
		}
		return base.Execute(ctx, st)
	}}
}

// CrashAposCheckpoints devolve um RunStore cujo Update falha depois de
// `commits` confirmações, simulando morte do processo no checkpoint; o
// estado já confirmado permanece intacto para recuperação.
func CrashAposCheckpoints(base wf.RunStore, commits int) wf.RunStore {
	return &storeComCrash{RunStore: base, restantes: commits}
}

type storeComCrash struct {
	wf.RunStore
	restantes int
}

func (s *storeComCrash) Update(ctx context.Context, run wf.Run) error {
	if s.restantes <= 0 {
		return fmt.Errorf("testkit: crash simulado no checkpoint do run %s", run.ID)
	}
	s.restantes--
	return s.RunStore.Update(ctx, run)
}
