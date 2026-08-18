package memory

import (
	"context"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Outbox guarda efeitos pendentes em memória com deduplicação por chave de
// idempotência, na mesma semântica exigida do adapter SQLite.
type Outbox struct {
	mu     sync.Mutex
	ordem  []string
	itens  map[string]workflow.EffectRecord
	vistos map[string]bool
}

// NewOutbox cria o outbox vazio.
func NewOutbox() *Outbox {
	return &Outbox{
		itens:  make(map[string]workflow.EffectRecord),
		vistos: make(map[string]bool),
	}
}

// Enqueue adiciona o efeito; ID já visto (pendente ou confirmado) é
// ignorado silenciosamente — reentrega é idempotente.
func (o *Outbox) Enqueue(_ context.Context, rec workflow.EffectRecord) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.vistos[rec.ID] {
		return nil
	}
	o.vistos[rec.ID] = true
	o.itens[rec.ID] = rec
	o.ordem = append(o.ordem, rec.ID)
	return nil
}

// Pending devolve os efeitos ainda não confirmados, na ordem de chegada.
func (o *Outbox) Pending(_ context.Context) ([]workflow.EffectRecord, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	pendentes := make([]workflow.EffectRecord, 0, len(o.itens))
	for _, id := range o.ordem {
		if rec, ok := o.itens[id]; ok {
			pendentes = append(pendentes, rec)
		}
	}
	return pendentes, nil
}

// Ack confirma o despacho do efeito; o ID permanece em vistos para que a
// reentrega continue deduplicada.
func (o *Outbox) Ack(_ context.Context, id string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.itens, id)
	return nil
}
