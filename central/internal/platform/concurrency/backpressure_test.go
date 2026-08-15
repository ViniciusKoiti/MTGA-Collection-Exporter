package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// TestPollBlockedSenderStaysBounded: with a slow consumer the poll
// producer blocks on the bounded channel instead of fetching ahead —
// the classic blocked-sender/slow-consumer proof.
func TestPollBlockedSenderStaysBounded(t *testing.T) {
	const buffer = 2
	var fetched, consumed atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, gctx := errgroup.WithContext(ctx)
	items := Poll(gctx, g, 5*time.Millisecond, buffer,
		func(context.Context) (int, bool, error) {
			return int(fetched.Add(1)), true, nil
		})
	g.Go(func() error {
		for range items {
			time.Sleep(10 * time.Millisecond) // slow consumer
			if consumed.Add(1) == 10 {
				cancel()
				return nil
			}
		}
		return nil
	})
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("wait: %v", err)
	}
	lag := fetched.Load() - consumed.Load()
	if lag > buffer+2 { // buffer plus the item in each hand
		t.Fatalf("blocked sender ran ahead: fetched %d consumed %d",
			fetched.Load(), consumed.Load())
	}
}
