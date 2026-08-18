package memory

import (
	"fmt"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// IDs gera RunIDs sequenciais e determinísticos para testes e harness.
type IDs struct {
	mu      sync.Mutex
	prefixo string
	proximo int
}

// NewIDs cria a fonte com o prefixo dado (ex.: "run").
func NewIDs(prefixo string) *IDs {
	return &IDs{prefixo: prefixo, proximo: 1}
}

// NewRunID devolve o próximo ID determinístico: prefixo-000001, ...
func (s *IDs) NewRunID() workflow.RunID {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := workflow.RunID(fmt.Sprintf("%s-%06d", s.prefixo, s.proximo))
	s.proximo++
	return id
}
