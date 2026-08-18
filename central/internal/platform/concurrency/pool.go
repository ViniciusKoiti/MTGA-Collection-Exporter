package concurrency

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// FlatPool consome `in` com `workers` goroutines, aplica fn e envia cada
// resultado individual em um canal limitado por `buffer`. O pool é o dono do
// canal de saída e o fecha somente quando todos os workers terminarem.
// Um erro de fn cancela o contexto do grupo, parando os demais estágios.
func FlatPool[In, Out any](
	ctx context.Context,
	g *errgroup.Group,
	in <-chan In,
	workers, buffer int,
	fn func(context.Context, In) ([]Out, error),
) <-chan Out {
	out := make(chan Out, buffer)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		g.Go(func() error {
			defer wg.Done()
			return runWorker(ctx, in, out, fn)
		})
	}
	g.Go(func() error {
		wg.Wait()
		close(out)
		return nil
	})
	return out
}

// Pool é o caso comum um-para-um sobre FlatPool.
func Pool[In, Out any](
	ctx context.Context,
	g *errgroup.Group,
	in <-chan In,
	workers, buffer int,
	fn func(context.Context, In) (Out, error),
) <-chan Out {
	return FlatPool(ctx, g, in, workers, buffer,
		func(ctx context.Context, v In) ([]Out, error) {
			r, err := fn(ctx, v)
			if err != nil {
				return nil, err
			}
			return []Out{r}, nil
		})
}

// runWorker processa itens até o canal de entrada fechar ou o contexto cancelar.
func runWorker[In, Out any](
	ctx context.Context,
	in <-chan In,
	out chan<- Out,
	fn func(context.Context, In) ([]Out, error),
) error {
	for {
		select {
		case v, ok := <-in:
			if !ok {
				return nil
			}
			results, err := fn(ctx, v)
			if err != nil {
				return err
			}
			for _, r := range results {
				if err := Send(ctx, out, r); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
