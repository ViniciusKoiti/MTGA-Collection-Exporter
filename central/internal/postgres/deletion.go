package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CompleteDeletion purges the installation's pseudonymous telemetry,
// blanks its secret material and stamps the deletion request with
// non-identifying completion evidence — all in ONE transaction. It is
// idempotent by design: reapplying a tombstone is a no-op.
func CompleteDeletion(ctx context.Context, pool *pgxpool.Pool,
	installationID string) (string, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("deletion.tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM accepted_events WHERE batch_id IN
		(SELECT id FROM telemetry_batches WHERE installation_id = $1)`,
		installationID); err != nil {
		return "", mapError("deletion.events", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM telemetry_batches
		WHERE installation_id = $1`, installationID); err != nil {
		return "", mapError("deletion.batches", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE installations SET token_hash = '',
		deletion_hash = '', revoked = TRUE WHERE id = $1`,
		installationID); err != nil {
		return "", mapError("deletion.blank", err)
	}
	proof := tombstoneProof(installationID)
	if _, err := tx.Exec(ctx, `UPDATE deletion_requests
		SET completed_at = now(), proof = $2
		WHERE installation_id = $1 AND completed_at IS NULL`,
		installationID, proof); err != nil {
		return "", mapError("deletion.stamp", err)
	}
	return proof, tx.Commit(ctx)
}

// ReapplyTombstones re-runs completion for every completed deletion —
// exactly what a backup restore must do, so purged data can never
// rise again from a backup.
func ReapplyTombstones(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	rows, err := pool.Query(ctx, `SELECT DISTINCT installation_id
		FROM deletion_requests WHERE completed_at IS NOT NULL`)
	if err != nil {
		return 0, mapError("deletion.tombstones", err)
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, mapError("deletion.scan", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, mapError("deletion.rows", err)
	}
	for _, id := range ids {
		if _, err := CompleteDeletion(ctx, pool, id); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

// tombstoneProof is one-way evidence: it proves THIS deletion ran
// without carrying the installation identity in the clear.
func tombstoneProof(installationID string) string {
	sum := sha256.Sum256([]byte("deletion-tombstone-v1:" + installationID))
	return hex.EncodeToString(sum[:8])
}
