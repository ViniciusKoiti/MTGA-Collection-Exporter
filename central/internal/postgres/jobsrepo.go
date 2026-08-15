package postgres

import (
	"context"
	"time"
)

// Job is the typed durable-job row (design decision 8).
type Job struct {
	ID             string
	Kind           string
	IdempotencyKey string
	Status         string
	Attempts       int
	Payload        []byte
	CreatedAt      time.Time
}

// JobsRepo owns the typed queries of the jobs table.
type JobsRepo struct {
	Q Querier
}

// Enqueue inserts the job; a duplicate idempotency key maps to the
// stable ErrDuplicate so transactional enqueue stays exactly-once.
func (r JobsRepo) Enqueue(ctx context.Context, job Job) error {
	_, err := r.Q.Exec(ctx, `INSERT INTO jobs (id, kind, idempotency_key, payload)
		VALUES ($1, $2, $3, COALESCE(NULLIF($4::text,'')::jsonb, '{}'::jsonb))`,
		job.ID, job.Kind, job.IdempotencyKey, string(job.Payload))
	return mapError("jobs.enqueue", err)
}

// Get loads one job by ID; absence is the stable ErrNotFound.
func (r JobsRepo) Get(ctx context.Context, id string) (Job, error) {
	var job Job
	var payload string
	err := r.Q.QueryRow(ctx, `SELECT id, kind, idempotency_key, status,
		attempts, payload::text, created_at FROM jobs WHERE id = $1`, id).
		Scan(&job.ID, &job.Kind, &job.IdempotencyKey, &job.Status,
			&job.Attempts, &payload, &job.CreatedAt)
	if err != nil {
		return Job{}, mapError("jobs.get", err)
	}
	job.Payload = []byte(payload)
	return job, nil
}

// ClaimNext leases the oldest pending job using FOR UPDATE SKIP LOCKED,
// so concurrent workers never claim the same job (groundwork for 6.1).
func (r JobsRepo) ClaimNext(ctx context.Context, owner string, lease time.Duration) (Job, error) {
	var job Job
	var payload string
	err := r.Q.QueryRow(ctx, `UPDATE jobs SET status = 'claimed',
		lease_owner = $1, lease_until = now() + $2::interval,
		attempts = attempts + 1, updated_at = now()
		WHERE id = (SELECT id FROM jobs
			WHERE status = 'pending' OR (status = 'claimed' AND lease_until < now())
			ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		RETURNING id, kind, idempotency_key, status, attempts, payload::text, created_at`,
		owner, lease.String()).
		Scan(&job.ID, &job.Kind, &job.IdempotencyKey, &job.Status,
			&job.Attempts, &payload, &job.CreatedAt)
	if err != nil {
		return Job{}, mapError("jobs.claim", err)
	}
	job.Payload = []byte(payload)
	return job, nil
}
