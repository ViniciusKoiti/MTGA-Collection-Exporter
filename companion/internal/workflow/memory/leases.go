package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

type lease struct {
	owner string
	until time.Time
}

// AcquireLease toma o run se estiver livre, expirado ou já for do dono.
func (s *RunStore) AcquireLease(
	_ context.Context,
	id workflow.RunID,
	owner string,
	agora, until time.Time,
) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, existe := s.runs[id]; !existe {
		return false, fmt.Errorf("memory: run %s não existe", id)
	}
	atual, possui := s.leases[id]
	if possui && atual.owner != owner && atual.until.After(agora) {
		return false, nil
	}
	s.leases[id] = lease{owner: owner, until: until}
	return true, nil
}

// ReleaseLease devolve o run; dono divergente é ignorado.
func (s *RunStore) ReleaseLease(_ context.Context, id workflow.RunID, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if atual, possui := s.leases[id]; possui && atual.owner == owner {
		delete(s.leases, id)
	}
	return nil
}

// RecoverableRuns lista runs ativos sem lease vigente.
func (s *RunStore) RecoverableRuns(_ context.Context, agora time.Time) ([]workflow.RunID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ids []workflow.RunID
	for id, run := range s.runs {
		if run.Status != workflow.RunActive {
			continue
		}
		if atual, possui := s.leases[id]; possui && atual.until.After(agora) {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// AbandonExpired encerra como abandonado todo run não terminal, sem lease
// vigente, parado desde antes de `cutoff`.
func (s *RunStore) AbandonExpired(_ context.Context, agora, cutoff time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	encerrados := 0
	for id, run := range s.runs {
		if run.Status.Terminal() || !run.UpdatedAt.Before(cutoff) {
			continue
		}
		if atual, possui := s.leases[id]; possui && atual.until.After(agora) {
			continue
		}
		run.Status = workflow.RunFailed
		run.Outcome = workflow.OutcomeAbandoned
		run.Version++
		run.UpdatedAt = agora
		s.runs[id] = run
		encerrados++
	}
	return encerrados, nil
}
