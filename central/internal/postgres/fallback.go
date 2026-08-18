package postgres

import (
	"context"
	"fmt"
)

// CurrentArtifactID answers which artifact clients receive right now.
func (a CatalogActivator) CurrentArtifactID(ctx context.Context) (string, error) {
	var id string
	err := a.Pool.QueryRow(ctx,
		`SELECT id FROM artifacts WHERE is_current`).Scan(&id)
	if err != nil {
		return "", mapError("activator.current", err)
	}
	return id, nil
}

// RollbackToPrevious re-activates the most recently published
// non-current artifact in ONE transaction, restoring the previous
// snapshot as the current one.
func (a CatalogActivator) RollbackToPrevious(ctx context.Context) error {
	tx, err := a.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("activator.rollback_tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var previous string
	err = tx.QueryRow(ctx, `SELECT id FROM artifacts
		WHERE NOT is_current AND published_at IS NOT NULL
		ORDER BY published_at DESC LIMIT 1`).Scan(&previous)
	if err != nil {
		return mapError("activator.previous", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE artifacts SET is_current = FALSE WHERE is_current`); err != nil {
		return mapError("activator.rollback_demote", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE artifacts SET is_current = TRUE WHERE id = $1`, previous); err != nil {
		return mapError("activator.rollback_promote", err)
	}
	return tx.Commit(ctx)
}
