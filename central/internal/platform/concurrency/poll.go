package concurrency

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

// Poll produz itens chamando fetch em um laço: cada item obtido é
// enviado imediatamente; quando não há trabalho, espera o próximo tick.
// É o produtor do poller de jobs (tarefa 6.2). Dono do canal devolvido.
func Poll[T any](
	ctx context.Context,
	g *errgroup.Group,
	every time.Duration,
	buffer int,
	fetch func(context.Context) (T, bool, error),
) <-chan T {
	out := make(chan T, buffer)
	g.Go(func() error {
		defer close(out)
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			item, found, err := fetch(ctx)
			if err != nil {
				return err
			}
			if found {
				if err := Send(ctx, out, item); err != nil {
					return err
				}
				continue // drain the backlog before sleeping
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	return out
}

// WithHeartbeat runs `work` while ticking `beat` at the given cadence;
// the first error from either side cancels the other. A failing beat
// (lost lease) aborts the work, so a fenced-out owner stops producing
// effects (tarefa 6.1/6.2).
func WithHeartbeat(
	ctx context.Context,
	every time.Duration,
	beat func(context.Context) error,
	work func(context.Context) error,
) error {
	g, gctx := errgroup.WithContext(ctx)
	done := make(chan struct{})
	g.Go(func() error {
		defer close(done)
		return work(gctx)
	})
	g.Go(func() error {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return nil
			case <-gctx.Done():
				return gctx.Err()
			case <-ticker.C:
				if err := beat(gctx); err != nil {
					return err
				}
			}
		}
	})
	return g.Wait()
}
