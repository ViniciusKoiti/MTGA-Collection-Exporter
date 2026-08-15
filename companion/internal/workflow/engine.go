package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Engine executa grafos registrados resolvendo apenas transições compiladas
// a partir de outcomes tipados (decisão 2 do design). Toda dependência chega
// por port na composição; o engine não conhece SQLite, Wails nem modelo.
type Engine struct {
	registry *Registry
	clock    Clock
	ids      IDSource
	store    RunStore
}

// NewEngine monta o engine com os ports obrigatórios.
func NewEngine(registry *Registry, clock Clock, ids IDSource, store RunStore) *Engine {
	return &Engine{registry: registry, clock: clock, ids: ids, store: store}
}

// Start cria o run, persiste o checkpoint inicial e executa até um outcome
// terminal, um limite ou cancelamento. O run devolvido reflete o estado
// final persistido; erros de limite são identificáveis com errors.Is.
func (e *Engine) Start(ctx context.Context, id Identity, initial State) (Run, error) {
	def, err := e.registry.Get(id)
	if err != nil {
		return Run{}, err
	}
	agora := e.clock.Now()
	run := Run{
		ID:        e.ids.NewRunID(),
		Graph:     id,
		Status:    RunActive,
		Current:   def.Initial,
		State:     initial,
		Version:   1,
		StartedAt: agora,
		UpdatedAt: agora,
	}
	if err := e.store.Create(ctx, run); err != nil {
		return Run{}, err
	}
	return e.loop(ctx, def, run)
}

// finish persiste o desfecho terminal com outcome estável e devolve o erro
// tipado correspondente (nil em sucesso).
func (e *Engine) finish(
	ctx context.Context,
	run Run,
	status RunStatus,
	outcome OutcomeCode,
	causa error,
) (Run, error) {
	run.Status = status
	run.Outcome = outcome
	run.Version++
	run.UpdatedAt = e.clock.Now()
	if err := e.store.Update(ctx, run); err != nil {
		return run, errors.Join(causa, err)
	}
	return run, causa
}

// registraStep grava a evidência da tentativa sem payloads (spec
// workflow-observability: nada de estado, entradas ou saídas no journal).
func (e *Engine) registraStep(
	ctx context.Context,
	run Run,
	indice int,
	outcome OutcomeCode,
	execErr error,
	inicio, fim time.Time,
) error {
	step := Step{
		Run:      run.ID,
		Index:    indice,
		Node:     run.Current,
		Attempt:  1,
		Outcome:  outcome,
		Started:  inicio,
		Finished: fim,
	}
	if execErr != nil {
		step.Err = string(OutcomeNodeError)
		if errors.Is(execErr, ErrLimitExceeded) {
			step.Err = string(OutcomeToolLimit)
		}
	}
	if err := e.store.AppendStep(ctx, step); err != nil {
		return fmt.Errorf("workflow: falha ao registrar step: %w", err)
	}
	return nil
}
