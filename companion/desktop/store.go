//go:build windows

package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/adapters/sqlitestore"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/homesvc"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

var (
	storeOnce sync.Once
	store     *sqlitestore.Store
	storeErr  error
)

// dataDir resolves where the app keeps its state: the user's config
// dir, or COMPANION_DATA_DIR when set — the runtime e2e suite uses
// that override so tests never touch real data.
func dataDir() (string, error) {
	if dir := os.Getenv("COMPANION_DATA_DIR"); dir != "" {
		return dir, nil
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(config, "MTGA-Companion"), nil
}

// collectionsStore lazily opens the local store under the data dir;
// an open failure surfaces as a view error state, never a crash.
func collectionsStore() (*sqlitestore.CollectionStore, error) {
	storeOnce.Do(func() {
		dir, err := dataDir()
		if err != nil {
			storeErr = err
			return
		}
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

// collections narrows the store to what the Home view needs.
func collections() (homesvc.SnapshotReader, error) {
	return collectionsStore()
}

// errorReader turns a store-open failure into an error view state.
type errorReader struct{ err error }

func (r errorReader) Latest(context.Context) (collection.Snapshot, bool, error) {
	return collection.Snapshot{}, false, r.err
}
