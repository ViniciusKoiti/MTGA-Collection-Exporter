package sqlitestore

import (
	"context"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// AppendStep adiciona a evidência da tentativa ao journal do run.
func (s *Store) AppendStep(ctx context.Context, step wf.Step) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO steps
		(run_id, idx, node, attempt, outcome, err, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		string(step.Run), step.Index, string(step.Node), step.Attempt,
		string(step.Outcome), step.Err,
		formataInstante(step.Started), formataInstante(step.Finished))
	return err
}

// Steps devolve as tentativas na ordem registrada (seq crescente).
func (s *Store) Steps(ctx context.Context, id wf.RunID) ([]wf.Step, error) {
	linhas, err := s.db.QueryContext(ctx, `SELECT run_id, idx, node, attempt,
		outcome, err, started_at, finished_at
		FROM steps WHERE run_id = ? ORDER BY seq`, string(id))
	if err != nil {
		return nil, err
	}
	defer func() { _ = linhas.Close() }()
	var steps []wf.Step
	for linhas.Next() {
		var step wf.Step
		var runID, node, outcome, inicio, fim string
		if err := linhas.Scan(&runID, &step.Index, &node, &step.Attempt,
			&outcome, &step.Err, &inicio, &fim); err != nil {
			return nil, err
		}
		step.Run = wf.RunID(runID)
		step.Node = wf.NodeID(node)
		step.Outcome = wf.OutcomeCode(outcome)
		if step.Started, err = parseInstante(inicio); err != nil {
			return nil, err
		}
		if step.Finished, err = parseInstante(fim); err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, linhas.Err()
}
