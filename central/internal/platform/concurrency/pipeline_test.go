package concurrency

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// pipelineDeQuadrados monta source -> pool -> reduce com os limites dados.
func pipelineDeQuadrados(ctx context.Context, itens []int, workers, buffer int) ([]int, error) {
	g, gctx := errgroup.WithContext(ctx)
	in := Source(gctx, g, itens, buffer)
	out := Pool(gctx, g, in, workers, buffer, func(_ context.Context, v int) (int, error) {
		return v * v, nil
	})
	var reduzido []int
	g.Go(func() error {
		var err error
		reduzido, err = Reduce(gctx, out, func(a, b int) bool { return a < b })
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return reduzido, nil
}

func TestPoolReduzDeterministicamente(t *testing.T) {
	itens := []int{9, 1, 7, 3, 5, 2, 8, 4, 6, 0}
	esperado := []int{0, 1, 4, 9, 16, 25, 36, 49, 64, 81}
	for range 20 { // várias execuções expõem ordem não determinística
		got, err := pipelineDeQuadrados(t.Context(), itens, 4, 2)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		for i, v := range esperado {
			if got[i] != v {
				t.Fatalf("ordem não determinística: %v", got)
			}
		}
	}
}

func TestErroCancelaPipelineInteiro(t *testing.T) {
	falha := errors.New("provedor caiu")
	g, gctx := errgroup.WithContext(t.Context())
	in := Source(gctx, g, make([]int, 1000), 1)
	out := Pool(gctx, g, in, 2, 1, func(_ context.Context, v int) (int, error) {
		return 0, falha
	})
	g.Go(func() error {
		_, err := Reduce(gctx, out, func(a, b int) bool { return a < b })
		return err
	})
	if err := g.Wait(); !errors.Is(err, falha) {
		t.Fatalf("primeiro erro deveria propagar, veio: %v", err)
	}
}

func TestCancelamentoDestravaSenderBloqueado(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g, gctx := errgroup.WithContext(ctx)
	// Ninguém consome: com buffer 1, os senders ficam bloqueados até o cancel.
	Source(gctx, g, make([]int, 100), 1)
	cancel()
	pronto := make(chan error, 1)
	go func() { pronto <- g.Wait() }()
	select {
	case err := <-pronto:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("esperava context.Canceled, veio: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("sender bloqueado não respeitou o cancelamento")
	}
}
