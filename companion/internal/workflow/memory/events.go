package memory

import (
	"context"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Events registra eventos em memória, na ordem de emissão, para asserções
// do harness e testes de observabilidade.
type Events struct {
	mu     sync.Mutex
	trilha []workflow.Event
}

// NewEvents cria o sink vazio.
func NewEvents() *Events {
	return &Events{}
}

// Emit registra o evento já validado pelo ValidatingSink do engine.
func (s *Events) Emit(_ context.Context, ev workflow.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trilha = append(s.trilha, ev)
	return nil
}

// Trilha devolve uma cópia dos eventos na ordem de emissão.
func (s *Events) Trilha() []workflow.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]workflow.Event(nil), s.trilha...)
}
