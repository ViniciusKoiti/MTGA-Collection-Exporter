package sqlitestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

var _ ports.TelemetryQueue = (*Store)(nil)

// AppendTelemetry persists one local product event in the durable
// outbox; it stays on the device until an ack covers its seq.
func (s *Store) AppendTelemetry(ctx context.Context, name string, at time.Time,
	attrs map[string]string) error {
	encoded, err := json.Marshal(attrs)
	if err != nil {
		return fmt.Errorf("sqlitestore: telemetry attrs: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO telemetry_outbox
		(name, at, attrs_json) VALUES (?, ?, ?)`,
		name, at.UTC().Format(time.RFC3339Nano), string(encoded))
	if err != nil {
		return fmt.Errorf("sqlitestore: telemetry append: %w", err)
	}
	return nil
}

// Pending lists unacked events in seq order, bounded by limite.
func (s *Store) Pending(ctx context.Context,
	limite int) ([]ports.TelemetryEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT seq, name, at, attrs_json
		FROM telemetry_outbox WHERE acked = 0 ORDER BY seq LIMIT ?`, limite)
	if err != nil {
		return nil, fmt.Errorf("sqlitestore: telemetry pending: %w", err)
	}
	defer rows.Close()
	var out []ports.TelemetryEvent
	for rows.Next() {
		var event ports.TelemetryEvent
		var at, attrs string
		if err := rows.Scan(&event.Seq, &event.Name, &at, &attrs); err != nil {
			return nil, fmt.Errorf("sqlitestore: telemetry scan: %w", err)
		}
		if event.At, err = time.Parse(time.RFC3339Nano, at); err != nil {
			return nil, fmt.Errorf("sqlitestore: telemetry at: %w", err)
		}
		if err := json.Unmarshal([]byte(attrs), &event.Attrs); err != nil {
			return nil, fmt.Errorf("sqlitestore: telemetry attrs: %w", err)
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

// Ack marks every event up to ateSeq as delivered; acknowledging the
// same seq twice is a no-op, which keeps retries idempotent.
func (s *Store) Ack(ctx context.Context, ateSeq int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE telemetry_outbox SET acked = 1
		WHERE acked = 0 AND seq <= ?`, ateSeq)
	if err != nil {
		return fmt.Errorf("sqlitestore: telemetry ack: %w", err)
	}
	return nil
}
