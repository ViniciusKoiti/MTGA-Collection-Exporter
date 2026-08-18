package workflow

import (
	"context"
	"sync"
)

// MetricKey é o único rótulo permitido nas métricas do runtime: tipo de
// grafo + outcome. Run, usuário, carta e deck jamais viram rótulo
// (tarefa 6.4 — baixa cardinalidade por construção).
type MetricKey struct {
	Kind    Kind
	Outcome OutcomeCode
}

// MetricValue agrega contagem e duração total em milissegundos.
type MetricValue struct {
	Count   int64
	TotalMS int64
}

// Metrics acumula contadores e durações em memória; exportadores reais
// (OpenTelemetry) leem o Snapshot.
type Metrics struct {
	mu      sync.Mutex
	valores map[MetricKey]MetricValue
}

// NewMetrics cria o agregador vazio.
func NewMetrics() *Metrics {
	return &Metrics{valores: make(map[MetricKey]MetricValue)}
}

func (m *Metrics) observa(ev Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	chave := MetricKey{Kind: ev.Graph.Kind, Outcome: ev.Outcome}
	valor := m.valores[chave]
	valor.Count++
	valor.TotalMS += ev.DurationMS
	m.valores[chave] = valor
}

// Snapshot devolve uma cópia estável dos agregados.
func (m *Metrics) Snapshot() map[MetricKey]MetricValue {
	m.mu.Lock()
	defer m.mu.Unlock()
	copia := make(map[MetricKey]MetricValue, len(m.valores))
	for chave, valor := range m.valores {
		copia[chave] = valor
	}
	return copia
}

// MetricsSink decora um EventSink agregando cada evento antes de repassar.
// Encadeia-se após o ValidatingSink, então só eventos íntegros contam.
type MetricsSink struct {
	Metrics *Metrics
	Next    EventSink
}

// Emit agrega e repassa ao próximo sink.
func (s MetricsSink) Emit(ctx context.Context, ev Event) error {
	s.Metrics.observa(ev)
	return s.Next.Emit(ctx, ev)
}
