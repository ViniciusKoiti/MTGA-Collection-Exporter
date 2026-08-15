package sqlitestore

import (
	"context"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// CollectionStore persists immutable collection snapshots on the same
// local database (OpenSpec introduce-agentic-go-companion, task 3.4):
// one transaction writes the snapshot row, every entry, every diagnostic
// and the latest flag — or nothing at all.
type CollectionStore struct {
	store *Store
}

var _ ports.SnapshotStore = (*CollectionStore)(nil)

// Collections exposes the snapshot repository backed by this database.
func (s *Store) Collections() *CollectionStore {
	return &CollectionStore{store: s}
}

// Save writes the snapshot transactionally. Snapshots are immutable: an
// existing ID fails the whole transaction and nothing is partially kept.
func (c *CollectionStore) Save(ctx context.Context, snap collection.Snapshot) error {
	tx, err := c.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO snapshots
		(id, schema, source, source_instance, observed_at, imported_at, is_latest)
		VALUES (?, ?, ?, ?, ?, ?, 1)`,
		string(snap.ID), snap.Schema, string(snap.Source), snap.SourceInstance,
		formataInstante(snap.ObservedAt), formataInstante(snap.ImportedAt)); err != nil {
		return fmt.Errorf("sqlitestore: snapshot %s is immutable or invalid: %w", snap.ID, err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE snapshots SET is_latest = 0 WHERE id != ?`, string(snap.ID)); err != nil {
		return err
	}
	for i, entry := range snap.Entries {
		if _, err := tx.ExecContext(ctx, `INSERT INTO snapshot_entries
			(snapshot_id, idx, printing, arena, oracle, name, set_code,
			 quantity, unresolved, raw)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			string(snap.ID), i, string(entry.Identity.Printing),
			int(entry.Identity.Arena), string(entry.Identity.Oracle),
			entry.Identity.Name, entry.Identity.Set, entry.Quantity,
			boolInt(entry.Unresolved), entry.Raw); err != nil {
			return err
		}
	}
	for i, diag := range snap.Diagnostics {
		if _, err := tx.ExecContext(ctx, `INSERT INTO snapshot_diagnostics
			(snapshot_id, idx, code, detail, severity) VALUES (?, ?, ?, ?, ?)`,
			string(snap.ID), i, diag.Code, diag.Detail, string(diag.Severity)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
