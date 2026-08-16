//go:build windows

package main

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/homesvc"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// mtgaRunning detects the game process without touching its memory.
func mtgaRunning(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "tasklist",
		"/FI", "IMAGENAME eq MTGA.exe", "/NH").Output()
	return err == nil && strings.Contains(string(out), "MTGA.exe")
}

// Home composes the Home view model against the real adapters.
func (a *App) Home() homesvc.Model {
	reader, err := collections()
	if err != nil {
		reader = errorReader{err: err}
	}
	deps := homesvc.Deps{
		Presence: mtgaRunning,
		Source: func(ctx context.Context) (collection.SourceKind, bool) {
			source, probeErr := clientProbe{}.Decide(ctx)
			if probeErr != nil {
				return collection.SourceJSONImport, false
			}
			return source, (sourceVerifier{}).Verify(ctx, source) == nil
		},
		Reader: reader,
		Now:    time.Now,
	}
	return homesvc.Compose(context.Background(), deps)
}
