package sqlitestore

import (
	"context"
	"fmt"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// CheckpointComEfeitos grava o checkpoint otimista e enfileira os efeitos
// do passo em UMA transação (tarefa 3.2): ou o novo estado e seus efeitos
// ficam visíveis juntos, ou nada fica. Conflito de versão desfaz tudo.
// IDs de efeito já vistos são ignorados (entrega at-least-once idempotente).
func (s *Store) CheckpointComEfeitos(
	ctx context.Context,
	run wf.Run,
	efeitos []wf.EffectRecord,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.updateEm(ctx, tx, run); err != nil {
		return err
	}
	for _, efeito := range efeitos {
		if efeito.ID == "" {
			return fmt.Errorf("sqlitestore: efeito sem chave de idempotência")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO outbox
			(id, run_id, effect, target, payload_hash, enqueued_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO NOTHING`,
			efeito.ID, string(efeito.Run), efeito.Preview.Effect,
			efeito.Preview.Target, efeito.Preview.PayloadHash,
			formataInstante(efeito.EnqueuedAt)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// EfeitosPendentes lista os efeitos não confirmados na ordem de chegada.
func (s *Store) EfeitosPendentes(ctx context.Context) ([]wf.EffectRecord, error) {
	linhas, err := s.db.QueryContext(ctx, `SELECT id, run_id, effect, target,
		payload_hash, enqueued_at FROM outbox WHERE acked = 0 ORDER BY enqueued_at, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = linhas.Close() }()
	var pendentes []wf.EffectRecord
	for linhas.Next() {
		var efeito wf.EffectRecord
		var runID, enfileirado string
		if err := linhas.Scan(&efeito.ID, &runID, &efeito.Preview.Effect,
			&efeito.Preview.Target, &efeito.Preview.PayloadHash, &enfileirado); err != nil {
			return nil, err
		}
		efeito.Run = wf.RunID(runID)
		if efeito.EnqueuedAt, err = parseInstante(enfileirado); err != nil {
			return nil, err
		}
		pendentes = append(pendentes, efeito)
	}
	return pendentes, linhas.Err()
}

// AckEfeito confirma o despacho; a linha permanece para deduplicação.
func (s *Store) AckEfeito(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE outbox SET acked = 1 WHERE id = ?`, id)
	return err
}
