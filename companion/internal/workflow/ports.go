package workflow

import (
	"context"
	"errors"
	"time"
)

// Clock abstrai o tempo para que o harness use relógio controlado.
type Clock interface {
	Now() time.Time
}

// IDSource abstrai a geração de IDs de run para execução determinística.
type IDSource interface {
	NewRunID() RunID
}

// RunStore persiste checkpoints e steps. Update é otimista: só aceita
// run.Version igual à versão armazenada + 1; conflito devolve
// ErrVersionConflict e o chamador decide recuperar ou abortar.
type RunStore interface {
	Create(ctx context.Context, run Run) error
	Update(ctx context.Context, run Run) error
	Get(ctx context.Context, id RunID) (Run, error)
	AppendStep(ctx context.Context, step Step) error
	Steps(ctx context.Context, id RunID) ([]Step, error)
}

// ErrVersionConflict indica escrita concorrente detectada pela versão
// otimista do checkpoint.
var ErrVersionConflict = errors.New("workflow: conflito de versão do run")

// Outcomes estáveis emitidos pelo engine quando um limite ou falha encerra
// o run; fazem parte do contrato observável e nunca mudam de texto.
const (
	OutcomeCancelled        OutcomeCode = "cancelled"
	OutcomeStepLimit        OutcomeCode = "limit_steps_exceeded"
	OutcomeDeadlineExceeded OutcomeCode = "limit_deadline_exceeded"
	OutcomeToolLimit        OutcomeCode = "limit_tool_calls_exceeded"
	OutcomePayloadLimit     OutcomeCode = "limit_payload_exceeded"
	OutcomeNodeError        OutcomeCode = "node_error"
	OutcomeUnmapped         OutcomeCode = "unmapped_outcome"
)
