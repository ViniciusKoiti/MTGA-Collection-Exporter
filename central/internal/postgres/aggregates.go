package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AggregateEvents rolls accepted events into hourly buckets and
// deletes the consumed rows IN THE SAME statement: counting and
// consuming are atomic, so no event can ever be counted twice. It
// returns the number of (metric, bucket) series touched.
func AggregateEvents(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx, `WITH consumed AS (
			DELETE FROM accepted_events e
			USING telemetry_batches b
			WHERE e.batch_id = b.id AND e.expires_at >= now()
			RETURNING e.name, date_trunc('hour', b.accepted_at) AS bucket
		)
		INSERT INTO aggregates (metric, bucket, value)
		SELECT name, bucket, count(*) FROM consumed GROUP BY name, bucket
		ON CONFLICT (metric, bucket)
		DO UPDATE SET value = aggregates.value + EXCLUDED.value`)
	if err != nil {
		return 0, mapError("aggregates.roll", err)
	}
	return tag.RowsAffected(), nil
}

// PurgeExpiredEvents drops events past their retention WITHOUT
// aggregating them — data kept too long is deleted, never mined.
func PurgeExpiredEvents(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx,
		`DELETE FROM accepted_events WHERE expires_at < now()`)
	if err != nil {
		return 0, mapError("aggregates.purge", err)
	}
	return tag.RowsAffected(), nil
}

// AggregateBucket is one non-identifying rollup row.
type AggregateBucket struct {
	Metric string
	Bucket time.Time
	Value  int64
}

// Totals is the ONLY read path of telemetry data: aggregates, never
// accepted events — queries below the rollup simply do not exist.
func Totals(ctx context.Context, q Querier, metric string) ([]AggregateBucket, error) {
	rows, err := q.Query(ctx, `SELECT metric, bucket, value FROM aggregates
		WHERE metric = $1 ORDER BY bucket`, metric)
	if err != nil {
		return nil, mapError("aggregates.totals", err)
	}
	defer rows.Close()
	var out []AggregateBucket
	for rows.Next() {
		var b AggregateBucket
		if err := rows.Scan(&b.Metric, &b.Bucket, &b.Value); err != nil {
			return nil, mapError("aggregates.scan", err)
		}
		out = append(out, b)
	}
	return out, mapError("aggregates.rows", rows.Err())
}
