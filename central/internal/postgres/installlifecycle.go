package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RotateToken swaps the credential hash as a compare-and-swap: the old
// hash must still match and the installation must not be revoked, so a
// stolen stale credential can never rotate itself back in.
func (r InstallationsRepo) RotateToken(ctx context.Context, id, oldHash,
	newHash string) error {
	tag, err := r.Q.Exec(ctx, `UPDATE installations SET token_hash = $3
		WHERE id = $1 AND token_hash = $2 AND NOT revoked`,
		id, oldHash, newHash)
	if err != nil {
		return mapError("installations.rotate", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Revoke disables the installation credential immediately.
func (r InstallationsRepo) Revoke(ctx context.Context, id string) error {
	tag, err := r.Q.Exec(ctx,
		`UPDATE installations SET revoked = TRUE WHERE id = $1`, id)
	if err != nil {
		return mapError("installations.revoke", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RequestDeletion verifies the deletion-secret hash and files the
// deletion request plus the revocation in ONE transaction; a wrong
// secret is indistinguishable from an unknown installation.
func RequestDeletion(ctx context.Context, pool *pgxpool.Pool, id,
	deletionHash string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("installations.deletion_tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var stored string
	err = tx.QueryRow(ctx,
		`SELECT deletion_hash FROM installations WHERE id = $1`, id).
		Scan(&stored)
	if err != nil {
		return mapError("installations.deletion_lookup", err)
	}
	if stored != deletionHash {
		return ErrNotFound
	}
	if _, err := tx.Exec(ctx, `INSERT INTO deletion_requests
		(installation_id) VALUES ($1)`, id); err != nil {
		return mapError("installations.deletion_file", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE installations SET revoked = TRUE WHERE id = $1`, id); err != nil {
		return mapError("installations.deletion_revoke", err)
	}
	return tx.Commit(ctx)
}
