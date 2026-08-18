package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig configures one process pool inside the reserved capacity.
type PoolConfig struct {
	DSN            string
	MaxConns       int           // pool ceiling for this process class
	AcquireTimeout time.Duration // waiting longer than this is saturation
	QueryTimeout   time.Duration // server-side statement_timeout
}

// NewPool builds a pgxpool honoring the budget: connection ceiling,
// acquisition timeout and a server-enforced statement timeout on every
// connection. Saturation surfaces as an acquire error, never as more
// connections.
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	if cfg.MaxConns < 1 || cfg.AcquireTimeout <= 0 || cfg.QueryTimeout <= 0 {
		return nil, fmt.Errorf("postgres: pool budget incomplete: %+v", cfg)
	}
	parsed, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres: pool dsn: %w", err)
	}
	parsed.MaxConns = int32(cfg.MaxConns)
	parsed.MinConns = 0
	parsed.ConnConfig.ConnectTimeout = cfg.AcquireTimeout
	parsed.ConnConfig.RuntimeParams["statement_timeout"] =
		fmt.Sprintf("%d", cfg.QueryTimeout.Milliseconds())
	pool, err := pgxpool.NewWithConfig(ctx, parsed)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// Acquire wraps pool.Acquire with the budgeted timeout so callers cannot
// wait forever on a saturated pool.
func Acquire(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) (*pgxpool.Conn, error) {
	acquireCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := pool.Acquire(acquireCtx)
	if err != nil {
		return nil, fmt.Errorf("postgres: pool saturated or unavailable: %w", err)
	}
	return conn, nil
}

// PoolHealth is the low-cardinality health snapshot for readiness and
// metrics: totals only, never per-query detail.
type PoolHealth struct {
	Total        int32
	Idle         int32
	Acquired     int32
	EmptyAcquire int64 // times a caller had to wait for a connection
}

// Health reads the pool statistics snapshot.
func Health(pool *pgxpool.Pool) PoolHealth {
	stat := pool.Stat()
	return PoolHealth{
		Total:        stat.TotalConns(),
		Idle:         stat.IdleConns(),
		Acquired:     stat.AcquiredConns(),
		EmptyAcquire: stat.EmptyAcquireCount(),
	}
}
