package sqlitestore

import (
	"context"
	"fmt"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// AcquireLease toma o run se estiver livre, expirado ou já for do dono;
// a disputa é resolvida pela condição do próprio UPDATE.
func (s *Store) AcquireLease(
	ctx context.Context,
	id wf.RunID,
	owner string,
	agora, until time.Time,
) (bool, error) {
	resultado, err := s.db.ExecContext(ctx, `UPDATE runs
		SET lease_owner = ?, lease_until = ?
		WHERE id = ? AND (lease_owner = '' OR lease_owner = ? OR lease_until < ?)`,
		owner, formataInstante(until), string(id), owner, formataInstante(agora))
	if err != nil {
		return false, err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return false, err
	}
	if afetadas > 0 {
		return true, nil
	}
	var existe int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM runs WHERE id = ?`, string(id)).Scan(&existe); err != nil {
		return false, err
	}
	if existe == 0 {
		return false, fmt.Errorf("sqlitestore: run %s não existe", id)
	}
	return false, nil
}

// ReleaseLease devolve o run; dono divergente é ignorado.
func (s *Store) ReleaseLease(ctx context.Context, id wf.RunID, owner string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE runs SET lease_owner = '', lease_until = ''
		WHERE id = ? AND lease_owner = ?`, string(id), owner)
	return err
}

// RecoverableRuns lista runs ativos sem lease vigente.
func (s *Store) RecoverableRuns(ctx context.Context, agora time.Time) ([]wf.RunID, error) {
	linhas, err := s.db.QueryContext(ctx, `SELECT id FROM runs
		WHERE status = ? AND (lease_owner = '' OR lease_until < ?) ORDER BY id`,
		string(wf.RunActive), formataInstante(agora))
	if err != nil {
		return nil, err
	}
	defer func() { _ = linhas.Close() }()
	var ids []wf.RunID
	for linhas.Next() {
		var id string
		if err := linhas.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, wf.RunID(id))
	}
	return ids, linhas.Err()
}

// AbandonExpired encerra como abandonado todo run não terminal, sem lease
// vigente, parado desde antes de `cutoff`.
func (s *Store) AbandonExpired(ctx context.Context, agora, cutoff time.Time) (int, error) {
	resultado, err := s.db.ExecContext(ctx, `UPDATE runs
		SET status = ?, outcome = ?, run_version = run_version + 1, updated_at = ?
		WHERE status IN (?, ?, ?) AND updated_at < ?
		  AND (lease_owner = '' OR lease_until < ?)`,
		string(wf.RunFailed), string(wf.OutcomeAbandoned), formataInstante(agora),
		string(wf.RunPending), string(wf.RunActive), string(wf.RunWaiting),
		formataInstante(cutoff), formataInstante(agora))
	if err != nil {
		return 0, err
	}
	afetadas, err := resultado.RowsAffected()
	return int(afetadas), err
}
