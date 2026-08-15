package testkit

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Controlled adapters (OpenSpec add-graph-workflow-harness, task 5.3):
// deterministic implementations of the PRODUCTION ports, driven only
// by scenario data. The controlled clock already exists as
// memory.Clock; these cover ID, planner, provider, process source,
// central API and local effect.

// ControlledIDs yields sequential IDs — never randomness in a run.
type ControlledIDs struct {
	Prefix string
	next   int
}

// New returns the next deterministic ID.
func (c *ControlledIDs) New() string {
	c.next++
	return fmt.Sprintf("%s-%04d", c.Prefix, c.next)
}

// ControlledModel is the planner behind ports.Model: it answers only
// from its script and refuses anything unscripted — a hallucinated
// plan cannot slip into a deterministic run.
type ControlledModel struct {
	Script map[string]string
}

var _ ports.Model = ControlledModel{}

// Respond answers the scripted reply for the exact context.
func (m ControlledModel) Respond(_ context.Context, contexto string) (string, error) {
	reply, ok := m.Script[contexto]
	if !ok {
		return "", fmt.Errorf("testkit: unscripted model context %q", contexto)
	}
	return reply, nil
}

// ControlledSource is the process-facing collection source: it serves
// one scenario observation through ports.CollectionSource.
type ControlledSource struct {
	Observation collection.Observation
	Err         error
}

var _ ports.CollectionSource = ControlledSource{}

// Kind reports the scenario source kind.
func (s ControlledSource) Kind() collection.SourceKind {
	return s.Observation.Source
}

// Observe serves the scripted observation or the scripted failure.
func (s ControlledSource) Observe(context.Context) (collection.Observation, error) {
	if s.Err != nil {
		return collection.Observation{}, s.Err
	}
	return s.Observation, nil
}

// ControlledCentral is the central API sink: it records batches and
// fails the first N sends so retry paths can be exercised.
type ControlledCentral struct {
	FailFirst int

	mu      sync.Mutex
	fails   int
	Eventos [][]ports.TelemetryEvent
}

var _ ports.TelemetrySink = (*ControlledCentral)(nil)

// SendBatch records the batch or fails while the budget lasts.
func (c *ControlledCentral) SendBatch(_ context.Context,
	lote []ports.TelemetryEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fails < c.FailFirst {
		c.fails++
		return errors.New("testkit: scripted central outage")
	}
	c.Eventos = append(c.Eventos, append([]ports.TelemetryEvent(nil), lote...))
	return nil
}
