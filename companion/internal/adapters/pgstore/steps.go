package pgstore

import (
	"context"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// AppendStep appends the attempt evidence to the run journal.
func (s *Store) AppendStep(ctx context.Context, step wf.Step) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO wf_steps
		(run_id, idx, node, attempt, outcome, err, started_at, finished_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		string(step.Run), step.Index, string(step.Node), step.Attempt,
		string(step.Outcome), step.Err, instant(step.Started), instant(step.Finished))
	return err
}

// Steps returns the attempts in journal order.
func (s *Store) Steps(ctx context.Context, id wf.RunID) ([]wf.Step, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT run_id, idx, node, attempt,
		outcome, err, started_at, finished_at
		FROM wf_steps WHERE run_id=$1 ORDER BY seq`, string(id))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var steps []wf.Step
	for rows.Next() {
		var step wf.Step
		var runID, node, outcome, started, finished string
		if err := rows.Scan(&runID, &step.Index, &node, &step.Attempt,
			&outcome, &step.Err, &started, &finished); err != nil {
			return nil, err
		}
		step.Run = wf.RunID(runID)
		step.Node = wf.NodeID(node)
		step.Outcome = wf.OutcomeCode(outcome)
		if step.Started, err = parseInstant(started); err != nil {
			return nil, err
		}
		if step.Finished, err = parseInstant(finished); err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

// reset truncates both tables; used by the contract suite to hand every
// subtest a clean store.
func (s *Store) reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE wf_runs, wf_steps`)
	return err
}
