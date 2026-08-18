package postgres

import (
	"context"
	"fmt"
	"time"
)

// Job lifecycle (OpenSpec add-central-go-platform, task 6.1): claimed
// jobs are owned through an expiring lease; the owner heartbeats while
// working, then finishes with a terminal outcome or hands the job back
// under the retry policy. A dead worker simply stops heartbeating and
// the expired lease makes the job claimable again.

// Terminal and retryable job states.
const (
	JobPending   = "pending"
	JobClaimed   = "claimed"
	JobSucceeded = "succeeded"
	JobFailed    = "failed"
)

// Heartbeat extends the lease while the owner still works. Losing the
// lease (expiry or reclaim by another worker) surfaces as ErrNotFound:
// the owner must stop producing effects immediately.
func (r JobsRepo) Heartbeat(ctx context.Context, id, owner string, lease time.Duration) error {
	tag, err := r.Q.Exec(ctx, `UPDATE jobs
		SET lease_until = now() + $3::interval, updated_at = now()
		WHERE id = $1 AND lease_owner = $2 AND status = 'claimed'
		  AND lease_until >= now()`,
		id, owner, lease.String())
	if err != nil {
		return mapError("jobs.heartbeat", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("jobs.heartbeat: lease lost: %w", ErrNotFound)
	}
	return nil
}

// Complete records the terminal success of an owned job.
func (r JobsRepo) Complete(ctx context.Context, id, owner string) error {
	tag, err := r.Q.Exec(ctx, `UPDATE jobs
		SET status = 'succeeded', lease_owner = '', lease_until = NULL,
		    updated_at = now()
		WHERE id = $1 AND lease_owner = $2 AND status = 'claimed'`, id, owner)
	if err != nil {
		return mapError("jobs.complete", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("jobs.complete: lease lost: %w", ErrNotFound)
	}
	return nil
}

// Fail applies the retry policy: below maxAttempts the job returns to
// pending for redelivery; at the budget it reaches the terminal failed
// state. The boolean reports whether the outcome was terminal.
func (r JobsRepo) Fail(ctx context.Context, id, owner string, maxAttempts int) (bool, error) {
	var status string
	err := r.Q.QueryRow(ctx, `UPDATE jobs
		SET status = CASE WHEN attempts >= $3 THEN 'failed' ELSE 'pending' END,
		    lease_owner = '', lease_until = NULL, updated_at = now()
		WHERE id = $1 AND lease_owner = $2 AND status = 'claimed'
		RETURNING status`, id, owner, maxAttempts).Scan(&status)
	if err != nil {
		return false, mapError("jobs.fail", err)
	}
	return status == JobFailed, nil
}
