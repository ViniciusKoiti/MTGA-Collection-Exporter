// Package lifecycle owns the ordered shutdown of the whole process
// (OpenSpec add-central-go-platform, task 7.3): fail readiness first,
// stop claims, drain workers, checkpoint, stop HTTP, close pools —
// always in that order, and pools always close.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"
)

// Hooks names each shutdown stage; every stage is mandatory because a
// forgotten stage is exactly how connections and jobs leak.
type Hooks struct {
	FailReadiness func()                          // stop attracting traffic
	StopClaims    func()                          // no new jobs claimed
	DrainWorkers  func(ctx context.Context) error // wait for in-flight jobs
	Checkpoint    func(ctx context.Context) error // persist progress
	StopHTTP      func(ctx context.Context) error // drain http server
	ClosePools    func()                          // release database conns
}

func (h Hooks) validate() error {
	if h.FailReadiness == nil || h.StopClaims == nil || h.DrainWorkers == nil ||
		h.Checkpoint == nil || h.StopHTTP == nil || h.ClosePools == nil {
		return errors.New("lifecycle: every shutdown stage is mandatory")
	}
	return nil
}

// SignalContext derives a context that ends on SIGINT or SIGTERM.
func SignalContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
}

// Run blocks until ctx is signalled, then executes the stages in
// order within the budget. A failing stage is recorded but never
// stops the later ones — pools close no matter what.
func Run(ctx context.Context, budget time.Duration, hooks Hooks) ([]string, error) {
	if err := hooks.validate(); err != nil {
		return nil, err
	}
	<-ctx.Done()
	stop, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	var done []string
	var errs []error
	step := func(name string, run func(context.Context) error) {
		if err := run(stop); err != nil {
			errs = append(errs, fmt.Errorf("lifecycle: %s: %w", name, err))
		}
		done = append(done, name)
	}
	hooks.FailReadiness()
	done = append(done, "fail_readiness")
	hooks.StopClaims()
	done = append(done, "stop_claims")
	step("drain_workers", hooks.DrainWorkers)
	step("checkpoint", hooks.Checkpoint)
	step("stop_http", hooks.StopHTTP)
	hooks.ClosePools()
	done = append(done, "close_pools")
	return done, errors.Join(errs...)
}
