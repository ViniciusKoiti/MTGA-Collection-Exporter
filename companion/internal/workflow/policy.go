package workflow

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"
)

// EffectPreview descreve exatamente o efeito que um nó pretende executar.
// A aprovação vale para este preview imutável; qualquer mudança de alvo ou
// payload gera novo hash e exige nova decisão (tarefa 2.5).
type EffectPreview struct {
	Effect      string // nome tipado do efeito (ex.: "export-write")
	Target      string // alvo exato (caminho, endpoint, identificador)
	PayloadHash string // hash do payload exato que será aplicado
}

// Hash identifica o preview de forma estável para aprovação e idempotência.
func (p EffectPreview) Hash() string {
	soma := sha256.Sum256([]byte(p.Effect + "\x00" + p.Target + "\x00" + p.PayloadHash))
	return fmt.Sprintf("%x", soma[:16])
}

// PolicyDecision é a decisão tipada da política para um preview.
type PolicyDecision string

// Decisões válidas de política.
const (
	PolicyAllow           PolicyDecision = "allow"
	PolicyRequireApproval PolicyDecision = "require_approval"
	PolicyDeny            PolicyDecision = "deny"
)

// Policy classifica o risco de um efeito para um grafo específico.
type Policy interface {
	Decide(ctx context.Context, graph Identity, preview EffectPreview) (PolicyDecision, error)
}

// Approval é o registro persistido de uma solicitação de aprovação.
type Approval struct {
	Run       RunID
	Hash      string // EffectPreview.Hash()
	Preview   EffectPreview
	Decided   bool
	Granted   bool
	ExpiresAt time.Time // aprovação concedida expira; zero = sem decisão
}

// ApprovalStore persiste solicitações e decisões de aprovação.
type ApprovalStore interface {
	Request(ctx context.Context, ap Approval) error
	Decide(ctx context.Context, run RunID, hash string, granted bool, expiresAt time.Time) error
	Get(ctx context.Context, run RunID, hash string) (Approval, bool, error)
}

// Erros tipados do fluxo de política e aprovação.
var (
	// ErrPolicyDenied encerra o run com outcome estável policy_denied.
	ErrPolicyDenied = errors.New("workflow: efeito negado pela política")

	// ErrApprovalPending pausa o run em waiting_approval até decisão.
	ErrApprovalPending = errors.New("workflow: efeito aguardando aprovação")

	// ErrApprovalExpired indica aprovação concedida cujo prazo venceu.
	ErrApprovalExpired = errors.New("workflow: aprovação expirada")
)

// Outcomes estáveis do fluxo de aprovação.
const (
	OutcomePolicyDenied    OutcomeCode = "policy_denied"
	OutcomeApprovalExpired OutcomeCode = "approval_expired"
)
