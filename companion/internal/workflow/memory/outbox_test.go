package memory

import (
	"testing"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func registro(id string) workflow.EffectRecord {
	return workflow.EffectRecord{
		ID:         id,
		Run:        "run-000001",
		Preview:    workflow.EffectPreview{Effect: "export-write", Target: "a", PayloadHash: "h"},
		EnqueuedAt: time.Unix(1_700_000_000, 0),
	}
}

func TestOutboxDeduplicaEConfirma(t *testing.T) {
	outbox := NewOutbox()
	ctx := t.Context()
	for range 3 { // reentrega do mesmo efeito é idempotente
		if err := outbox.Enqueue(ctx, registro("efeito-1")); err != nil {
			t.Fatalf("enqueue: %v", err)
		}
	}
	_ = outbox.Enqueue(ctx, registro("efeito-2"))
	pendentes, _ := outbox.Pending(ctx)
	if len(pendentes) != 2 || pendentes[0].ID != "efeito-1" || pendentes[1].ID != "efeito-2" {
		t.Fatalf("pendentes inesperados: %+v", pendentes)
	}
	if err := outbox.Ack(ctx, "efeito-1"); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if err := outbox.Enqueue(ctx, registro("efeito-1")); err != nil {
		t.Fatalf("reenqueue pós-ack: %v", err)
	}
	pendentes, _ = outbox.Pending(ctx)
	if len(pendentes) != 1 || pendentes[0].ID != "efeito-2" {
		t.Fatalf("ack + dedupe deveriam deixar só efeito-2: %+v", pendentes)
	}
}
