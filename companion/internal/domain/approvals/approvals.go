// Package approvals define os modelos de aprovação e auditoria do
// companion (tarefa 2.2 do OpenSpec introduce-agentic-go-companion):
// tokens de uso único vinculados aos argumentos exatos, e registros de
// auditoria redigidos. Só stdlib.
package approvals

import (
	"fmt"
	"time"
)

// Token autoriza UMA execução de UMA ferramenta com argumentos exatos
// (hash); expira e não é reutilizável (decisão de política do design).
type Token struct {
	ID        string
	Tool      string
	ArgsHash  string
	ExpiresAt time.Time
	UsedAt    time.Time // zero enquanto não usado
}

// NewToken valida os invariantes do token.
func NewToken(id, tool, argsHash string, expiresAt time.Time) (Token, error) {
	if id == "" || tool == "" || argsHash == "" {
		return Token{}, fmt.Errorf("approvals: token incompleto")
	}
	if expiresAt.IsZero() {
		return Token{}, fmt.Errorf("approvals: token sem expiração")
	}
	return Token{ID: id, Tool: tool, ArgsHash: argsHash, ExpiresAt: expiresAt}, nil
}

// Usavel informa se o token autoriza a execução agora, para a ferramenta
// e argumentos exatos apresentados.
func (t Token) Usavel(agora time.Time, tool, argsHash string) error {
	if !t.UsedAt.IsZero() {
		return fmt.Errorf("approvals: token %s já usado", t.ID)
	}
	if agora.After(t.ExpiresAt) {
		return fmt.Errorf("approvals: token %s expirado", t.ID)
	}
	if t.Tool != tool || t.ArgsHash != argsHash {
		return fmt.Errorf("approvals: token %s não cobre %s com estes argumentos", t.ID, tool)
	}
	return nil
}

// Consumir marca o uso único; devolve o token consumido imutavelmente.
func (t Token) Consumir(agora time.Time) Token {
	t.UsedAt = agora
	return t
}

// AuditOutcome é o desfecho estável de uma interação de ferramenta.
type AuditOutcome string

// Desfechos válidos de auditoria.
const (
	AuditRequested AuditOutcome = "requested"
	AuditApproved  AuditOutcome = "approved"
	AuditExecuted  AuditOutcome = "executed"
	AuditFailed    AuditOutcome = "failed"
	AuditDenied    AuditOutcome = "denied"
)

// AuditRecord é o registro redigido de uma interação: nunca carrega
// argumentos crus, payloads, caminhos ou credenciais — só códigos.
type AuditRecord struct {
	Correlation string
	Tool        string
	ArgsHash    string
	Outcome     AuditOutcome
	At          time.Time
}
