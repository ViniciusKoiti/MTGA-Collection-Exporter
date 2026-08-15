package sqlitestore

import (
	"context"
	"time"
)

// RetentionReport resume o que a política de retenção removeu.
type RetentionReport struct {
	RunsRemovidos    int
	EventosRemovidos int
}

// ApplyRetention aplica retenção por idade e por tamanho (tarefa 6.6):
//
//   - runs TERMINAIS atualizados antes de `corteIdade` são removidos com
//     seus steps, eventos, aprovações e efeitos já confirmados;
//   - se o journal exceder `maxEventos`, os mais antigos além do teto são
//     removidos, mas somente de runs terminais.
//
// Checkpoints de runs ativos/aguardando nunca são tocados: o checkpoint é
// autoritativo e a retenção não pode quebrar recuperação.
func (s *Store) ApplyRetention(
	ctx context.Context,
	corteIdade time.Time,
	maxEventos int,
) (RetentionReport, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RetentionReport{}, err
	}
	defer func() { _ = tx.Rollback() }()
	relatorio := RetentionReport{}

	terminais := `SELECT id FROM runs
		WHERE status IN ('succeeded', 'failed', 'cancelled') AND updated_at < ?`
	for _, comando := range []string{
		`DELETE FROM steps WHERE run_id IN (` + terminais + `)`,
		`DELETE FROM events WHERE run_id IN (` + terminais + `)`,
		`DELETE FROM approvals WHERE run_id IN (` + terminais + `)`,
		`DELETE FROM outbox WHERE acked = 1 AND run_id IN (` + terminais + `)`,
	} {
		if _, err := tx.ExecContext(ctx, comando, formataInstante(corteIdade)); err != nil {
			return RetentionReport{}, err
		}
	}
	removidos, err := tx.ExecContext(ctx,
		`DELETE FROM runs WHERE status IN ('succeeded', 'failed', 'cancelled')
		 AND updated_at < ?`, formataInstante(corteIdade))
	if err != nil {
		return RetentionReport{}, err
	}
	linhas, err := removidos.RowsAffected()
	if err != nil {
		return RetentionReport{}, err
	}
	relatorio.RunsRemovidos = int(linhas)

	excedente, err := tx.ExecContext(ctx, `DELETE FROM events WHERE seq IN (
		SELECT e.seq FROM events e
		JOIN runs r ON r.id = e.run_id
		WHERE r.status IN ('succeeded', 'failed', 'cancelled')
		ORDER BY e.seq
		LIMIT max(0, (SELECT COUNT(*) FROM events) - ?))`, maxEventos)
	if err != nil {
		return RetentionReport{}, err
	}
	if linhas, err = excedente.RowsAffected(); err != nil {
		return RetentionReport{}, err
	}
	relatorio.EventosRemovidos = int(linhas)
	return relatorio, tx.Commit()
}
