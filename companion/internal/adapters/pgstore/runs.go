package pgstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Create inserts the initial checkpoint; an existing run is an error.
func (s *Store) Create(ctx context.Context, run wf.Run) error {
	state, err := json.Marshal(run.State)
	if err != nil {
		return fmt.Errorf("pgstore: state not serializable: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO wf_runs
		(id, graph_kind, graph_version, status, current_node, state_json,
		 outcome, run_version, started_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		string(run.ID), string(run.Graph.Kind), run.Graph.Version, string(run.Status),
		string(run.Current), string(state), string(run.Outcome), run.Version,
		instant(run.StartedAt), instant(run.UpdatedAt))
	if err != nil {
		return fmt.Errorf("pgstore: create %s: %w", run.ID, err)
	}
	return nil
}

// Update applies the checkpoint only when the version is stored+1; the
// condition lives in the UPDATE itself, so the swap is atomic.
func (s *Store) Update(ctx context.Context, run wf.Run) error {
	state, err := json.Marshal(run.State)
	if err != nil {
		return fmt.Errorf("pgstore: state not serializable: %w", err)
	}
	result, err := s.db.ExecContext(ctx, `UPDATE wf_runs SET
		status=$1, current_node=$2, state_json=$3, outcome=$4,
		run_version=$5, updated_at=$6
		WHERE id=$7 AND run_version=$8`,
		string(run.Status), string(run.Current), string(state), string(run.Outcome),
		run.Version, instant(run.UpdatedAt), string(run.ID), run.Version-1)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: run %s at version %d", wf.ErrVersionConflict, run.ID, run.Version)
	}
	return nil
}

// Get returns the current checkpoint.
func (s *Store) Get(ctx context.Context, id wf.RunID) (wf.Run, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, graph_kind, graph_version,
		status, current_node, state_json, outcome, run_version, started_at, updated_at
		FROM wf_runs WHERE id=$1`, string(id))
	var run wf.Run
	var kind, status, current, state, outcome, started, updated string
	err := row.Scan(&run.ID, &kind, &run.Graph.Version, &status, &current,
		&state, &outcome, &run.Version, &started, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return wf.Run{}, fmt.Errorf("pgstore: run %s not found", id)
	}
	if err != nil {
		return wf.Run{}, err
	}
	run.Graph.Kind = wf.Kind(kind)
	run.Status = wf.RunStatus(status)
	run.Current = wf.NodeID(current)
	run.Outcome = wf.OutcomeCode(outcome)
	if err := json.Unmarshal([]byte(state), &run.State); err != nil {
		return wf.Run{}, fmt.Errorf("pgstore: corrupted state: %w", err)
	}
	if run.StartedAt, err = parseInstant(started); err != nil {
		return wf.Run{}, err
	}
	if run.UpdatedAt, err = parseInstant(updated); err != nil {
		return wf.Run{}, err
	}
	return run, nil
}
