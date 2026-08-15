package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Approvals guarda solicitações e decisões de aprovação em memória.
type Approvals struct {
	mu    sync.Mutex
	itens map[chaveAprovacao]workflow.Approval
}

type chaveAprovacao struct {
	run  workflow.RunID
	hash string
}

// NewApprovals cria o store vazio.
func NewApprovals() *Approvals {
	return &Approvals{itens: make(map[chaveAprovacao]workflow.Approval)}
}

// Request registra a solicitação pendente; não sobrescreve decisão existente.
func (s *Approvals) Request(_ context.Context, ap workflow.Approval) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	chave := chaveAprovacao{ap.Run, ap.Hash}
	if existente, ok := s.itens[chave]; ok && existente.Decided {
		return fmt.Errorf("memory: aprovação %s/%s já decidida", ap.Run, ap.Hash)
	}
	s.itens[chave] = ap
	return nil
}

// Decide grava a decisão do usuário para a solicitação pendente.
func (s *Approvals) Decide(
	_ context.Context,
	run workflow.RunID,
	hash string,
	granted bool,
	expiresAt time.Time,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	chave := chaveAprovacao{run, hash}
	ap, ok := s.itens[chave]
	if !ok {
		return fmt.Errorf("memory: aprovação %s/%s inexistente", run, hash)
	}
	ap.Decided = true
	ap.Granted = granted
	ap.ExpiresAt = expiresAt
	s.itens[chave] = ap
	return nil
}

// Get devolve a solicitação e se ela existe.
func (s *Approvals) Get(
	_ context.Context,
	run workflow.RunID,
	hash string,
) (workflow.Approval, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ap, ok := s.itens[chaveAprovacao{run, hash}]
	return ap, ok, nil
}
