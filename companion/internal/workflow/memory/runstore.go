package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// RunStore guarda checkpoints e steps em memória com a mesma semântica de
// versão otimista exigida do adapter SQLite (contrato RunStore).
type RunStore struct {
	mu     sync.Mutex
	runs   map[workflow.RunID]workflow.Run
	steps  map[workflow.RunID][]workflow.Step
	leases map[workflow.RunID]lease
}

// NewRunStore cria o store vazio.
func NewRunStore() *RunStore {
	return &RunStore{
		runs:   make(map[workflow.RunID]workflow.Run),
		steps:  make(map[workflow.RunID][]workflow.Step),
		leases: make(map[workflow.RunID]lease),
	}
}

// Create insere o checkpoint inicial; run já existente é erro.
func (s *RunStore) Create(_ context.Context, run workflow.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, existe := s.runs[run.ID]; existe {
		return fmt.Errorf("memory: run %s já existe", run.ID)
	}
	s.runs[run.ID] = run
	return nil
}

// Update aplica o checkpoint somente se a versão for a armazenada + 1.
func (s *RunStore) Update(_ context.Context, run workflow.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	atual, existe := s.runs[run.ID]
	if !existe {
		return fmt.Errorf("memory: run %s não existe", run.ID)
	}
	if run.Version != atual.Version+1 {
		return fmt.Errorf("%w: run %s esperava %d, veio %d",
			workflow.ErrVersionConflict, run.ID, atual.Version+1, run.Version)
	}
	s.runs[run.ID] = run
	return nil
}

// Get devolve o checkpoint corrente.
func (s *RunStore) Get(_ context.Context, id workflow.RunID) (workflow.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, existe := s.runs[id]
	if !existe {
		return workflow.Run{}, fmt.Errorf("memory: run %s não existe", id)
	}
	return run, nil
}

// AppendStep adiciona a evidência da tentativa ao journal do run.
func (s *RunStore) AppendStep(_ context.Context, step workflow.Step) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.steps[step.Run] = append(s.steps[step.Run], step)
	return nil
}

// Steps devolve as tentativas na ordem registrada.
func (s *RunStore) Steps(_ context.Context, id workflow.RunID) ([]workflow.Step, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]workflow.Step(nil), s.steps[id]...), nil
}
