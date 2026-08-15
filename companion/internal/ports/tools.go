package ports

import (
	"context"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
)

// Clock abstrai o tempo para aplicação e harness determinístico.
type Clock interface {
	Now() time.Time
}

// Clipboard escreve texto na área de transferência do usuário; é um
// efeito local que exige aprovação quando pedido pelo assistente.
type Clipboard interface {
	Write(ctx context.Context, texto string) error
}

// ApprovalService emite e resgata tokens de aprovação de uso único
// vinculados à ferramenta e ao hash exato dos argumentos.
type ApprovalService interface {
	Request(ctx context.Context, tool, argsHash string) (approvals.Token, error)
	Redeem(ctx context.Context, tokenID, tool, argsHash string) error
}

// AuditLog registra interações redigidas e as expõe para a visão de
// atividade do assistente.
type AuditLog interface {
	Append(ctx context.Context, rec approvals.AuditRecord) error
	Recent(ctx context.Context, limite int) ([]approvals.AuditRecord, error)
}

// Model é o provedor opcional do assistente. O desktop determinístico
// funciona sem ele; o contexto enviado já chega minimizado e redigido.
type Model interface {
	Respond(ctx context.Context, contexto string) (string, error)
}
