package workflow

import (
	"context"
	"time"
)

// EffectRecord é a intenção durável de executar um efeito exatamente uma
// vez do ponto de vista do consumidor: o ID é a chave de idempotência e a
// entrega é at-least-once (design, decisão 3).
type EffectRecord struct {
	ID         string // chave de idempotência do efeito
	Run        RunID
	Preview    EffectPreview
	EnqueuedAt time.Time
}

// Outbox persiste efeitos junto do checkpoint para despacho posterior.
// Enqueue com ID repetido é ignorado silenciosamente (idempotente);
// Ack remove o registro após confirmação do executor.
type Outbox interface {
	Enqueue(ctx context.Context, rec EffectRecord) error
	Pending(ctx context.Context) ([]EffectRecord, error)
	Ack(ctx context.Context, id string) error
}
