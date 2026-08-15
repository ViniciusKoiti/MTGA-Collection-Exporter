package sqlitestore

import (
	"context"
	"encoding/json"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Emit grava o evento no journal; o INSERT único é atômico e a ordem vem
// do seq autoincrementado (tarefa 6.2). O Store satisfaz wf.EventSink e é
// sempre encadeado atrás do ValidatingSink do engine.
func (s *Store) Emit(ctx context.Context, ev wf.Event) error {
	attrs := []byte("{}")
	if len(ev.Attrs) > 0 {
		serializado, err := json.Marshal(ev.Attrs)
		if err != nil {
			return err
		}
		attrs = serializado
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO events
		(schema, run_id, step, graph_kind, graph_version, correlation,
		 causation, at, outcome, duration_ms, attrs_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ev.Schema, string(ev.Run), ev.Step, string(ev.Graph.Kind), ev.Graph.Version,
		ev.Correlation, ev.Causation, formataInstante(ev.At), string(ev.Outcome),
		ev.DurationMS, string(attrs))
	return err
}

// Timeline devolve os eventos do run após o cursor, em ordem de journal,
// limitados por página; o segundo retorno é o cursor da próxima página
// (0 quando não há mais eventos).
func (s *Store) Timeline(
	ctx context.Context,
	id wf.RunID,
	aposSeq int64,
	limite int,
) ([]wf.Event, int64, error) {
	linhas, err := s.db.QueryContext(ctx, `SELECT seq, schema, run_id, step,
		graph_kind, graph_version, correlation, causation, at, outcome,
		duration_ms, attrs_json
		FROM events WHERE run_id = ? AND seq > ? ORDER BY seq LIMIT ?`,
		string(id), aposSeq, limite)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = linhas.Close() }()
	var eventos []wf.Event
	var cursor int64
	for linhas.Next() {
		var ev wf.Event
		var seq int64
		var runID, kind, instante, outcome, attrs string
		if err := linhas.Scan(&seq, &ev.Schema, &runID, &ev.Step, &kind,
			&ev.Graph.Version, &ev.Correlation, &ev.Causation, &instante,
			&outcome, &ev.DurationMS, &attrs); err != nil {
			return nil, 0, err
		}
		ev.Run = wf.RunID(runID)
		ev.Graph.Kind = wf.Kind(kind)
		ev.Outcome = wf.OutcomeCode(outcome)
		if ev.At, err = parseInstante(instante); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal([]byte(attrs), &ev.Attrs); err != nil {
			return nil, 0, err
		}
		eventos = append(eventos, ev)
		cursor = seq
	}
	if err := linhas.Err(); err != nil {
		return nil, 0, err
	}
	if len(eventos) < limite {
		cursor = 0 // não há próxima página
	}
	return eventos, cursor, nil
}
