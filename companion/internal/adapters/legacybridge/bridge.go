// Package legacybridge runs the legacy Python memory scanner as an
// EXPLICIT user compatibility command (OpenSpec
// introduce-agentic-go-companion, task 3.9): the Go process never opens
// MTGA memory itself — it invokes the isolated legacy pipeline, waits
// with a deadline, and imports the export it produced through the frozen
// legacy contract. The bridge is not an agent tool: the policy matrix
// classifies it prohibited for the assistant, and the exclusion test in
// this package proves it.
package legacybridge

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/legacyjson"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// exportName is the file the legacy pipeline writes into OutputDir.
const exportName = "mtga_collection.json"

// Bridge is a CollectionSource backed by the legacy Python scanner.
type Bridge struct {
	Command   []string // executable + args, injected by composition
	Env       []string // extra environment for the command
	OutputDir string   // where the legacy pipeline writes its export
	Timeout   time.Duration
	Clock     ports.Clock
}

var _ ports.CollectionSource = (*Bridge)(nil)

// Kind identifies observations as coming from the legacy bridge.
func (b *Bridge) Kind() collection.SourceKind { return collection.SourceLegacyBridge }

// Observe runs the legacy command under a deadline and imports the
// export it wrote; the observation is re-labeled with the bridge kind so
// provenance is never confused with a plain JSON import.
func (b *Bridge) Observe(ctx context.Context) (collection.Observation, error) {
	if len(b.Command) == 0 {
		return collection.Observation{}, fmt.Errorf("legacybridge: no command configured")
	}
	runCtx := ctx
	if b.Timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, b.Timeout)
		defer cancel()
	}
	command := exec.CommandContext(runCtx, b.Command[0], b.Command[1:]...)
	command.Env = append(command.Environ(), b.Env...)
	if output, err := command.CombinedOutput(); err != nil {
		return collection.Observation{}, fmt.Errorf(
			"legacybridge: scanner failed (%d output bytes): %w", len(output), err)
	}
	source := &legacyjson.Source{
		Path:  filepath.Join(b.OutputDir, exportName),
		Clock: b.Clock,
	}
	imported, err := source.Observe(ctx)
	if err != nil {
		return collection.Observation{}, fmt.Errorf(
			"legacybridge: scanner finished but export is unusable: %w", err)
	}
	return collection.NewObservation(collection.SourceLegacyBridge,
		imported.SourceInstance, imported.ObservedAt,
		imported.Quantities, imported.Diagnostics)
}
