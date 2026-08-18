package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// executorIdempotente conta chamadas brutas e aplica cada ID uma única vez,
// como o contrato do outbox exige.
type executorIdempotente struct {
	chamadas  int
	aplicados map[string]int
	falhas    int // primeiras N execuções falham (crash antes do efeito)
}

func (e *executorIdempotente) Execute(_ context.Context, rec wf.EffectRecord) error {
	e.chamadas++
	if e.falhas > 0 {
		e.falhas--
		return errors.New("fronteira indisponível")
	}
	if e.aplicados == nil {
		e.aplicados = map[string]int{}
	}
	if e.aplicados[rec.ID] == 0 {
		e.aplicados[rec.ID] = 1
	}
	return nil
}

// outboxSemAck simula morte entre Execute e Ack: o primeiro Ack falha.
type outboxSemAck struct {
	*memory.Outbox
	falhasAck int
}

func (o *outboxSemAck) Ack(ctx context.Context, id string) error {
	if o.falhasAck > 0 {
		o.falhasAck--
		return errors.New("crash simulado antes do ack")
	}
	return o.Outbox.Ack(ctx, id)
}

func registroEfeito(id string) wf.EffectRecord {
	return wf.EffectRecord{ID: id, Run: "run-000001",
		Preview:    wf.EffectPreview{Effect: "export-write", Target: "a", PayloadHash: "h"},
		EnqueuedAt: time.Unix(1_700_000_000, 0)}
}

func TestDispatchConfirmaEDeduplicaReentrega(t *testing.T) {
	outbox := memory.NewOutbox()
	_ = outbox.Enqueue(t.Context(), registroEfeito("e1"))
	_ = outbox.Enqueue(t.Context(), registroEfeito("e2"))
	exec := &executorIdempotente{}
	if n, err := wf.Dispatch(t.Context(), outbox, exec); err != nil || n != 2 {
		t.Fatalf("despacho: %d (%v)", n, err)
	}
	if n, err := wf.Dispatch(t.Context(), outbox, exec); err != nil || n != 0 {
		t.Fatalf("segundo despacho deveria ser vazio: %d (%v)", n, err)
	}
	if exec.chamadas != 2 || len(exec.aplicados) != 2 {
		t.Fatalf("execuções inesperadas: %+v", exec)
	}
}

func TestCrashAntesDoEfeitoMantemPendencia(t *testing.T) {
	outbox := memory.NewOutbox()
	_ = outbox.Enqueue(t.Context(), registroEfeito("e1"))
	exec := &executorIdempotente{falhas: 1}
	if _, err := wf.Dispatch(t.Context(), outbox, exec); err == nil {
		t.Fatal("falha da fronteira deveria interromper o despacho")
	}
	if n, err := wf.Dispatch(t.Context(), outbox, exec); err != nil || n != 1 {
		t.Fatalf("reentrega deveria confirmar: %d (%v)", n, err)
	}
	if exec.aplicados["e1"] != 1 {
		t.Fatalf("efeito deveria aplicar exatamente uma vez: %+v", exec.aplicados)
	}
}

func TestCrashEntreEfeitoEAckReentregaSemDuplicar(t *testing.T) {
	outbox := &outboxSemAck{Outbox: memory.NewOutbox(), falhasAck: 1}
	_ = outbox.Enqueue(t.Context(), registroEfeito("e1"))
	exec := &executorIdempotente{}
	if _, err := wf.Dispatch(t.Context(), outbox, exec); err == nil {
		t.Fatal("crash antes do ack deveria interromper o despacho")
	}
	if n, err := wf.Dispatch(t.Context(), outbox, exec); err != nil || n != 1 {
		t.Fatalf("reentrega pós-crash deveria confirmar: %d (%v)", n, err)
	}
	if exec.chamadas != 2 || exec.aplicados["e1"] != 1 {
		t.Fatalf("reexecução sim, duplicação não: %+v", exec)
	}
}
