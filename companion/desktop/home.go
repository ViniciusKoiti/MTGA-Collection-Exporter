//go:build windows

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/sqlitestore"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/homesvc"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

var (
	storeOnce sync.Once
	store     *sqlitestore.Store
	storeErr  error
)

// collections lazily opens the local store under the user's config
// dir; an open failure surfaces as the Home error state, never a
// crash.
func collections() (homesvc.SnapshotReader, error) {
	storeOnce.Do(func() {
		config, err := os.UserConfigDir()
		if err != nil {
			storeErr = err
			return
		}
		dir := filepath.Join(config, "MTGA-Companion")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			storeErr = err
			return
		}
		store, storeErr = sqlitestore.Open(filepath.Join(dir, "companion.db"))
	})
	if storeErr != nil {
		return nil, storeErr
	}
	return store.Collections(), nil
}

// errorReader turns a store-open failure into the Home error state.
type errorReader struct{ err error }

func (r errorReader) Latest(context.Context) (collection.Snapshot, bool, error) {
	return collection.Snapshot{}, false, r.err
}

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
