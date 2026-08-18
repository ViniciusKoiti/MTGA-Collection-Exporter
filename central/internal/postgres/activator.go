package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/ports"
)

// CatalogActivator is the transactional Activator adapter (OpenSpec
// add-central-go-platform, task 4.5): activation either fully promotes
// the new manifest or leaves the previous one current — never a state
// in between.
type CatalogActivator struct {
	Pool       *pgxpool.Pool
	SourceID   string
	SchemaName string
}

var _ ports.Activator = CatalogActivator{}

// Activate records snapshot plus artifact and swaps the current flag
// in ONE transaction; any failure keeps the previous manifest current.
func (a CatalogActivator) Activate(ctx context.Context,
	man catalog.Manifest) error {
	tx, err := a.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("activator.tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `INSERT INTO snapshots
		(id, source_id, schema_name) VALUES ($1, $2, $3)`,
		man.Object.Key, a.SourceID, a.SchemaName); err != nil {
		return mapError("activator.snapshot", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO artifacts
		(id, snapshot_id, object_key, sha256, size_bytes, signed_by,
		 is_current, published_at)
		VALUES ($1, $1, $1, $2, $3, $4, TRUE, now())`,
		man.Object.Key, man.Object.SHA256, man.Object.Size,
		man.KeyID); err != nil {
		return mapError("activator.artifact", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE artifacts SET is_current = FALSE
		WHERE is_current AND id <> $1`, man.Object.Key); err != nil {
		return mapError("activator.demote", err)
	}
	return tx.Commit(ctx)
}
