package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TelemetryEvent is one allowlisted event inside an accepted batch.
type TelemetryEvent struct {
	Name  string
	Attrs map[string]string
}

// TelemetryBatch is the idempotent unit of ingestion: the pair
// (installation, sequence) and the batch ID are both unique.
type TelemetryBatch struct {
	ID             string
	InstallationID string
	Sequence       int64
	Events         []TelemetryEvent
}

// IngestBatch persists the batch and its events in ONE transaction:
// either the batch and every event exist, or nothing does. A replayed
// sequence or batch ID maps to ErrDuplicate.
func IngestBatch(ctx context.Context, pool *pgxpool.Pool,
	batch TelemetryBatch, retention time.Duration) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("telemetry.ingest_tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `INSERT INTO telemetry_batches
		(id, installation_id, sequence) VALUES ($1, $2, $3)`,
		batch.ID, batch.InstallationID, batch.Sequence); err != nil {
		return mapError("telemetry.batch", err)
	}
	expires := time.Now().Add(retention)
	for _, event := range batch.Events {
		attrs, err := json.Marshal(event.Attrs)
		if err != nil {
			return fmt.Errorf("telemetry.attrs: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO accepted_events
			(batch_id, name, attrs, expires_at) VALUES ($1, $2, $3, $4)`,
			batch.ID, event.Name, attrs, expires); err != nil {
			return mapError("telemetry.event", err)
		}
	}
	return tx.Commit(ctx)
}
