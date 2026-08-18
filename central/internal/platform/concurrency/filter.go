package concurrency

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Filter forwards the items keep approves and drops the rest. keep
// runs in the single filter goroutine so its side effects (such as
// quarantining) need no extra synchronization; an error cancels the
// whole run through the errgroup.
func Filter[T any](ctx context.Context, g *errgroup.Group, in <-chan T,
	buffer int, keep func(context.Context, T) (bool, error)) <-chan T {
	out := make(chan T, buffer)
	g.Go(func() error {
		defer close(out)
		for item := range in {
			ok, err := keep(ctx, item)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			select {
			case out <- item:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})
	return out
}
