package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Get loads one snapshot with entries and diagnostics in insert order.
func (c *CollectionStore) Get(ctx context.Context, id collection.SnapshotID) (collection.Snapshot, error) {
	row := c.store.db.QueryRowContext(ctx, `SELECT id, schema, source, source_instance,
		observed_at, imported_at FROM snapshots WHERE id = ?`, string(id))
	var snap collection.Snapshot
	var source, observed, imported string
	err := row.Scan(&snap.ID, &snap.Schema, &source, &snap.SourceInstance,
		&observed, &imported)
	if errors.Is(err, sql.ErrNoRows) {
		return collection.Snapshot{}, fmt.Errorf("sqlitestore: snapshot %s not found", id)
	}
	if err != nil {
		return collection.Snapshot{}, err
	}
	snap.Source = collection.SourceKind(source)
	if snap.ObservedAt, err = parseInstante(observed); err != nil {
		return collection.Snapshot{}, err
	}
	if snap.ImportedAt, err = parseInstante(imported); err != nil {
		return collection.Snapshot{}, err
	}
	if snap.Entries, err = c.entriesOf(ctx, id); err != nil {
		return collection.Snapshot{}, err
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT code, detail, severity
		FROM snapshot_diagnostics WHERE snapshot_id = ? ORDER BY idx`, string(id))
	if err != nil {
		return collection.Snapshot{}, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var diag collection.Diagnostic
		var severity string
		if err := rows.Scan(&diag.Code, &diag.Detail, &severity); err != nil {
			return collection.Snapshot{}, err
		}
		diag.Severity = collection.Severity(severity)
		snap.Diagnostics = append(snap.Diagnostics, diag)
	}
	return snap, rows.Err()
}

func (c *CollectionStore) entriesOf(ctx context.Context, id collection.SnapshotID) ([]collection.Entry, error) {
	rows, err := c.store.db.QueryContext(ctx, `SELECT printing, arena, oracle, name,
		set_code, quantity, unresolved, raw
		FROM snapshot_entries WHERE snapshot_id = ? ORDER BY idx`, string(id))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var entries []collection.Entry
	for rows.Next() {
		var entry collection.Entry
		var printing, oracle string
		var arena, unresolved int
		if err := rows.Scan(&printing, &arena, &oracle, &entry.Identity.Name,
			&entry.Identity.Set, &entry.Quantity, &unresolved, &entry.Raw); err != nil {
			return nil, err
		}
		entry.Identity.Printing = collection.PrintingID(printing)
		entry.Identity.Arena = collection.ArenaID(arena)
		entry.Identity.Oracle = collection.OracleID(oracle)
		entry.Unresolved = unresolved == 1
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// Latest returns the snapshot flagged as current, if any.
func (c *CollectionStore) Latest(ctx context.Context) (collection.Snapshot, bool, error) {
	var id string
	err := c.store.db.QueryRowContext(ctx,
		`SELECT id FROM snapshots WHERE is_latest = 1`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return collection.Snapshot{}, false, nil
	}
	if err != nil {
		return collection.Snapshot{}, false, err
	}
	snap, err := c.Get(ctx, collection.SnapshotID(id))
	return snap, err == nil, err
}
