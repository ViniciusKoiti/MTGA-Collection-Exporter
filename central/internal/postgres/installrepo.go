package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Installation is the typed installation row.
type Installation struct {
	ID           string
	TokenHash    string
	DeletionHash string
	Revoked      bool
}

// InstallationsRepo owns the typed queries of installations and consent.
type InstallationsRepo struct {
	Q Querier
}

// Enroll inserts the installation; duplicates map to ErrDuplicate.
func (r InstallationsRepo) Enroll(ctx context.Context, inst Installation) error {
	_, err := r.Q.Exec(ctx, `INSERT INTO installations (id, token_hash, deletion_hash)
		VALUES ($1, $2, $3)`, inst.ID, inst.TokenHash, inst.DeletionHash)
	return mapError("installations.enroll", err)
}

// ByID loads one installation; absence is the stable ErrNotFound.
func (r InstallationsRepo) ByID(ctx context.Context, id string) (Installation, error) {
	var inst Installation
	err := r.Q.QueryRow(ctx, `SELECT id, token_hash, deletion_hash, revoked
		FROM installations WHERE id = $1`, id).
		Scan(&inst.ID, &inst.TokenHash, &inst.DeletionHash, &inst.Revoked)
	if err != nil {
		return Installation{}, mapError("installations.by_id", err)
	}
	return inst, nil
}

// RecordConsent appends a consent receipt for the installation.
func (r InstallationsRepo) RecordConsent(ctx context.Context, installationID,
	purpose string, version int) error {
	_, err := r.Q.Exec(ctx, `INSERT INTO consent_receipts
		(installation_id, purpose, version) VALUES ($1, $2, $3)`,
		installationID, purpose, version)
	return mapError("installations.consent", err)
}

// EnrollWithConsent runs enrollment and the first consent receipt in ONE
// transaction: either the installation exists with its receipt, or
// neither does (task 2.4: transactional repository adapters).
func EnrollWithConsent(ctx context.Context, pool *pgxpool.Pool,
	inst Installation, purpose string, version int) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("installations.enroll_tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	repo := InstallationsRepo{Q: tx}
	if err := repo.Enroll(ctx, inst); err != nil {
		return err
	}
	if err := repo.RecordConsent(ctx, inst.ID, purpose, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
