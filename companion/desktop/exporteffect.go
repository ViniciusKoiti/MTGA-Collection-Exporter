//go:build windows

package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/compatibility"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// exportWriter executes the export-write effect: it re-projects the
// export and writes it ONLY while the payload hash still matches the
// approved preview — a drifted collection voids the approval instead
// of silently exporting something the user never saw.
type exportWriter struct {
	dir string
}

func (w exportWriter) Execute(ctx context.Context, rec wf.EffectRecord) error {
	snapshots, err := collectionsStore()
	if err != nil {
		return err
	}
	snap, ok, err := snapshots.Latest(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("export effect: no snapshot to export")
	}
	format := ports.ExportFormat(filepath.Ext(rec.Preview.Target)[1:])
	payload, err := (compatibility.Exporter{}).Export(ctx, snap, format)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	if fmt.Sprintf("%x", sum) != rec.Preview.PayloadHash {
		return fmt.Errorf(
			"export effect: payload drifted; the approval no longer covers it")
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.dir, rec.Preview.Target),
		payload, 0o644)
}

// dispatchExports drains the outbox through the writer and answers
// where the export landed.
func (a *App) dispatchExports(ctx context.Context,
	stack *exportStack) (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	exportDir := filepath.Join(dir, "exports")
	if _, err := wf.Dispatch(ctx, stack.outbox,
		exportWriter{dir: exportDir}); err != nil {
		return "", err
	}
	return exportDir, nil
}
