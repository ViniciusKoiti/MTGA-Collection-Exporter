package sqlitestore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Previous returns the snapshot imported right before the latest one,
// backing the Home delta (homesvc.HistoryReader). With fewer than two
// snapshots there is no previous — that is a normal state, not an
// error.
func (c *CollectionStore) Previous(ctx context.Context) (collection.Snapshot, bool, error) {
	var id string
	err := c.store.db.QueryRowContext(ctx, `SELECT id FROM snapshots
		ORDER BY imported_at DESC, id DESC LIMIT 1 OFFSET 1`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return collection.Snapshot{}, false, nil
	}
	if err != nil {
		return collection.Snapshot{}, false, err
	}
	snap, err := c.Get(ctx, collection.SnapshotID(id))
	if err != nil {
		return collection.Snapshot{}, false, err
	}
	return snap, true, nil
}
