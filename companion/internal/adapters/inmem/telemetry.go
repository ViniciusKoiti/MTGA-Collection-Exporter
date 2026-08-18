package inmem

import (
	"context"
	"errors"
	"sync"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Consent devolve o estado fixo de consentimento do cenário.
type Consent struct {
	Estado ports.ConsentState
}

var _ ports.ConsentStore = (*Consent)(nil)

func (c *Consent) Current(context.Context) (ports.ConsentState, error) {
	return c.Estado, nil
}

// TelemetryQueue é a fila local durável em memória.
type TelemetryQueue struct {
	mu     sync.Mutex
	Itens  []ports.TelemetryEvent
	ackAte int
}

var _ ports.TelemetryQueue = (*TelemetryQueue)(nil)

// Pending devolve até `limite` eventos ainda não confirmados, em ordem.
func (q *TelemetryQueue) Pending(_ context.Context, limite int) ([]ports.TelemetryEvent, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var pendentes []ports.TelemetryEvent
	for _, ev := range q.Itens {
		if ev.Seq > q.ackAte {
			pendentes = append(pendentes, ev)
			if len(pendentes) == limite {
				break
			}
		}
	}
	return pendentes, nil
}

// Ack confirma todos os eventos até o seq dado.
func (q *TelemetryQueue) Ack(_ context.Context, ateSeq int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if ateSeq > q.ackAte {
		q.ackAte = ateSeq
	}
	return nil
}

// PendentesAgora conta os não confirmados (asserções de teste).
func (q *TelemetryQueue) PendentesAgora() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	total := 0
	for _, ev := range q.Itens {
		if ev.Seq > q.ackAte {
			total++
		}
	}
	return total
}

// TelemetrySink registra lotes enviados e falha as primeiras N chamadas.
type TelemetrySink struct {
	mu     sync.Mutex
	Falhas int // chamadas que ainda falharão
	Lotes  [][]ports.TelemetryEvent
}

var _ ports.TelemetrySink = (*TelemetrySink)(nil)

func (s *TelemetrySink) SendBatch(_ context.Context, lote []ports.TelemetryEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Falhas > 0 {
		s.Falhas--
		return errors.New("inmem: central indisponível")
	}
	copia := append([]ports.TelemetryEvent(nil), lote...)
	s.Lotes = append(s.Lotes, copia)
	return nil
}
