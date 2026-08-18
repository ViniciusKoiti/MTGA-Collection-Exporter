// Package memory fornece adapters in-memory dos ports do runtime para
// testes unitários e para o harness determinístico (tarefa 2.6 do OpenSpec
// add-graph-workflow-harness). Somente biblioteca padrão.
package memory

import (
	"sync"
	"time"
)

// Clock é um relógio controlado: só avança quando o teste manda.
type Clock struct {
	mu    sync.Mutex
	agora time.Time
}

// NewClock inicia o relógio no instante dado.
func NewClock(inicio time.Time) *Clock {
	return &Clock{agora: inicio}
}

// Now devolve o instante corrente controlado.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.agora
}

// Advance move o relógio adiante; nunca retrocede.
func (c *Clock) Advance(d time.Duration) {
	if d < 0 {
		panic("memory: relógio controlado não retrocede")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.agora = c.agora.Add(d)
}
