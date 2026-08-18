package lifecycle

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

func hooksRecording(order *[]string, failDrain bool) Hooks {
	mark := func(name string) func() {
		return func() { *order = append(*order, name) }
	}
	stage := func(name string, fail bool) func(context.Context) error {
		return func(context.Context) error {
			*order = append(*order, name)
			if fail {
				return errors.New(name + " failed")
			}
			return nil
		}
	}
	return Hooks{
		FailReadiness: mark("fail_readiness"),
		StopClaims:    mark("stop_claims"),
		DrainWorkers:  stage("drain_workers", failDrain),
		Checkpoint:    stage("checkpoint", false),
		StopHTTP:      stage("stop_http", false),
		ClosePools:    mark("close_pools"),
	}
}

var wantOrder = []string{"fail_readiness", "stop_claims", "drain_workers",
	"checkpoint", "stop_http", "close_pools"}

// TestShutdownRunsEveryStageInOrder: the signal triggers the exact
// documented sequence.
func TestShutdownRunsEveryStageInOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the "signal"
	var order []string
	done, err := Run(ctx, time.Second, hooksRecording(&order, false))
	if err != nil {
		t.Fatalf("clean shutdown: %v", err)
	}
	if !slices.Equal(order, wantOrder) || !slices.Equal(done, wantOrder) {
		t.Fatalf("stage order broken: %v / %v", order, done)
	}
}

// TestFailingStageNeverStopsPoolClosure: drain fails, yet checkpoint,
// http stop and pool closure all still run and the error surfaces.
func TestFailingStageNeverStopsPoolClosure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var order []string
	done, err := Run(ctx, time.Second, hooksRecording(&order, true))
	if err == nil {
		t.Fatal("the drain failure must surface")
	}
	if !slices.Equal(order, wantOrder) {
		t.Fatalf("failure must not skip stages: %v", order)
	}
	if done[len(done)-1] != "close_pools" {
		t.Fatalf("pools must always close: %v", done)
	}
}

func TestShutdownRefusesMissingStages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	hooks := hooksRecording(new([]string), false)
	hooks.Checkpoint = nil
	if _, err := Run(ctx, time.Second, hooks); err == nil {
		t.Fatal("a missing stage must be refused")
	}
}
