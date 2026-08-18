package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Create insere o checkpoint inicial; run já existente é erro.
func (s *Store) Create(ctx context.Context, run wf.Run) error {
	estado, err := serializaEstado(run.State)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO runs
		(id, graph_kind, graph_version, status, current_node, state_json,
		 outcome, run_version, started_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(run.ID), string(run.Graph.Kind), run.Graph.Version, string(run.Status),
		string(run.Current), estado, string(run.Outcome), run.Version,
		formataInstante(run.StartedAt), formataInstante(run.UpdatedAt))
	if err != nil {
		return fmt.Errorf("sqlitestore: create %s: %w", run.ID, err)
	}
	return nil
}

// Update aplica o checkpoint somente se a versão for a armazenada + 1;
// a condição vive no próprio UPDATE, então a troca é atômica.
func (s *Store) Update(ctx context.Context, run wf.Run) error {
	return s.updateEm(ctx, s.db, run)
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (s *Store) updateEm(ctx context.Context, ex executor, run wf.Run) error {
	estado, err := serializaEstado(run.State)
	if err != nil {
		return err
	}
	resultado, err := ex.ExecContext(ctx, `UPDATE runs SET
		status = ?, current_node = ?, state_json = ?, outcome = ?,
		run_version = ?, updated_at = ?
		WHERE id = ? AND run_version = ?`,
		string(run.Status), string(run.Current), estado, string(run.Outcome),
		run.Version, formataInstante(run.UpdatedAt), string(run.ID), run.Version-1)
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return fmt.Errorf("%w: run %s na versão %d", wf.ErrVersionConflict, run.ID, run.Version)
	}
	return nil
}

// Get devolve o checkpoint corrente.
func (s *Store) Get(ctx context.Context, id wf.RunID) (wf.Run, error) {
	linha := s.db.QueryRowContext(ctx, `SELECT id, graph_kind, graph_version,
		status, current_node, state_json, outcome, run_version, started_at, updated_at
		FROM runs WHERE id = ?`, string(id))
	var run wf.Run
	var kind, status, atual, estado, outcome, inicio, atualizacao string
	err := linha.Scan(&run.ID, &kind, &run.Graph.Version, &status, &atual,
		&estado, &outcome, &run.Version, &inicio, &atualizacao)
	if errors.Is(err, sql.ErrNoRows) {
		return wf.Run{}, fmt.Errorf("sqlitestore: run %s não existe", id)
	}
	if err != nil {
		return wf.Run{}, err
	}
	run.Graph.Kind = wf.Kind(kind)
	run.Status = wf.RunStatus(status)
	run.Current = wf.NodeID(atual)
	run.Outcome = wf.OutcomeCode(outcome)
	if run.State, err = desserializaEstado(estado); err != nil {
		return wf.Run{}, err
	}
	if run.StartedAt, err = parseInstante(inicio); err != nil {
		return wf.Run{}, err
	}
	if run.UpdatedAt, err = parseInstante(atualizacao); err != nil {
		return wf.Run{}, err
	}
	return run, nil
}
