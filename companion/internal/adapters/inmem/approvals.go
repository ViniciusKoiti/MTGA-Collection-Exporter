package inmem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/approvals"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// validadeToken é o prazo padrão de um token de aprovação em memória.
const validadeToken = 5 * time.Minute

// ApprovalService emite e resgata tokens de uso único em memória, usando
// as regras do domínio (approvals.Token).
type ApprovalService struct {
	mu      sync.Mutex
	clock   ports.Clock
	proximo int
	tokens  map[string]approvals.Token
}

var _ ports.ApprovalService = (*ApprovalService)(nil)

// NewApprovalService cria o serviço com o relógio dado.
func NewApprovalService(clock ports.Clock) *ApprovalService {
	return &ApprovalService{clock: clock, proximo: 1, tokens: make(map[string]approvals.Token)}
}

// Request emite um token vinculado à ferramenta e ao hash exato.
func (s *ApprovalService) Request(_ context.Context, tool, argsHash string) (approvals.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := fmt.Sprintf("tok-%06d", s.proximo)
	s.proximo++
	token, err := approvals.NewToken(id, tool, argsHash, s.clock.Now().Add(validadeToken))
	if err != nil {
		return approvals.Token{}, err
	}
	s.tokens[id] = token
	return token, nil
}

// Redeem consome o token se ele cobrir exatamente a ferramenta e os
// argumentos; qualquer divergência, expiração ou reuso é negado.
func (s *ApprovalService) Redeem(_ context.Context, tokenID, tool, argsHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.tokens[tokenID]
	if !ok {
		return fmt.Errorf("inmem: token %s inexistente", tokenID)
	}
	if err := token.Usavel(s.clock.Now(), tool, argsHash); err != nil {
		return err
	}
	s.tokens[tokenID] = token.Consumir(s.clock.Now())
	return nil
}
