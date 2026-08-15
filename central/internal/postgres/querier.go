package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is satisfied by both *pgxpool.Pool and pgx.Tx, so every
// repository method runs equally inside or outside a transaction
// (task 2.4: hand-owned typed queries instead of generated ones — the
// design allows either; contract suites keep them from drifting).
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Stable repository errors: callers branch on these, never on SQLSTATE
// strings or driver messages.
var (
	ErrNotFound  = errors.New("postgres: not found")
	ErrDuplicate = errors.New("postgres: duplicate")
)

// mapError converts driver errors into the stable vocabulary while
// preserving context cancellation as-is.
func mapError(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, ErrNotFound)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return fmt.Errorf("%s: %w", op, ErrDuplicate)
	}
	return fmt.Errorf("%s: %w", op, err)
}
