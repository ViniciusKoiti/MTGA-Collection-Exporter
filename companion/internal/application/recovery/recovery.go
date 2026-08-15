// Package recovery rebuilds the local database from a compatibility
// export (OpenSpec introduce-agentic-go-companion, task 7.2): when the
// database is lost, the last written mtga_collection.json — read through
// the frozen legacy contract — becomes a fresh authoritative snapshot.
package recovery

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/apperr"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/application/normalize"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// FromExport observes the export through the given source (the legacy
// JSON adapter in production), re-normalizes it against the catalog and
// commits the result as a new immutable snapshot in the store. The
// recovered snapshot carries a diagnostic marking its provenance.
func FromExport(
	ctx context.Context,
	source ports.CollectionSource,
	catalog ports.Catalog,
	store ports.SnapshotStore,
	clock ports.Clock,
) (collection.Snapshot, error) {
	observation, err := source.Observe(ctx)
	if err != nil {
		return collection.Snapshot{}, apperr.New(apperr.CodeSourceUnavailable,
			"recovery.from_export", err)
	}
	result, err := normalize.Observation(ctx, catalog, observation)
	if err != nil {
		return collection.Snapshot{}, apperr.New(apperr.CodeCatalogUnavailable,
			"recovery.from_export", err)
	}
	now := clock.Now()
	diagnostics := append(result.Diagnostics, collection.Diagnostic{
		Code: "recovered_from_export", Detail: string(observation.Source),
		Severity: collection.SeverityInfo,
	})
	// The source instance may be a local path; only its hash reaches the ID.
	instance := sha256.Sum256([]byte(observation.SourceInstance))
	snapshot, err := collection.NewSnapshot(
		collection.SnapshotID(fmt.Sprintf("recovered-%d-%x", now.Unix(), instance[:4])),
		observation, now, result.Entries, diagnostics)
	if err != nil {
		return collection.Snapshot{}, apperr.New(apperr.CodeValidationFailed,
			"recovery.from_export", err)
	}
	if err := store.Save(ctx, snapshot); err != nil {
		return collection.Snapshot{}, apperr.New(apperr.CodeInternal,
			"recovery.from_export", err)
	}
	return snapshot, nil
}
