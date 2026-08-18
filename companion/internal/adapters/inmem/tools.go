package inmem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Clock é um relógio controlado para o perfil determinístico.
type Clock struct {
	mu    sync.Mutex
	agora time.Time
}

var _ ports.Clock = (*Clock)(nil)

// NewClock inicia o relógio no instante dado.
func NewClock(inicio time.Time) *Clock { return &Clock{agora: inicio} }

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.agora
}

// Advance move o relógio adiante.
func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.agora = c.agora.Add(d)
}

// Clipboard registra o que seria copiado, sem tocar o sistema.
type Clipboard struct {
	mu     sync.Mutex
	Textos []string
}

var _ ports.Clipboard = (*Clipboard)(nil)

func (c *Clipboard) Write(_ context.Context, texto string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Textos = append(c.Textos, texto)
	return nil
}

// AuditLog acumula registros redigidos em ordem.
type AuditLog struct {
	mu     sync.Mutex
	trilha []approvals.AuditRecord
}

var _ ports.AuditLog = (*AuditLog)(nil)

func (a *AuditLog) Append(_ context.Context, rec approvals.AuditRecord) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.trilha = append(a.trilha, rec)
	return nil
}

func (a *AuditLog) Recent(_ context.Context, limite int) ([]approvals.AuditRecord, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	inicio := max(0, len(a.trilha)-limite)
	return append([]approvals.AuditRecord(nil), a.trilha[inicio:]...), nil
}

// FakeModel é o provedor de assistente do perfil determinístico: eco fixo.
type FakeModel struct{}

var _ ports.Model = FakeModel{}

func (FakeModel) Respond(_ context.Context, contexto string) (string, error) {
	return fmt.Sprintf("resposta-fake(%d bytes de contexto)", len(contexto)), nil
}
