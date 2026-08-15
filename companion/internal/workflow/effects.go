package workflow

import (
	"context"
	"fmt"
)

// effectGate media todo efeito de um run: consulta a política e o estado de
// aprovação antes de liberar. Nós nunca falam com Policy diretamente.
type effectGate struct {
	policy    Policy
	approvals ApprovalStore
	clock     Clock
	run       RunID
	graph     Identity
}

type effectGateKey struct{}

func withEffectGate(ctx context.Context, g *effectGate) context.Context {
	return context.WithValue(ctx, effectGateKey{}, g)
}

// RequestEffect é chamado pelo nó com o preview exato do efeito pretendido.
// nil libera a execução; ErrApprovalPending pausa o run com segurança;
// ErrPolicyDenied encerra com outcome estável. Sem gate no contexto
// (engine ausente), todo efeito é negado por segurança.
func RequestEffect(ctx context.Context, preview EffectPreview) error {
	gate, ok := ctx.Value(effectGateKey{}).(*effectGate)
	if !ok {
		return fmt.Errorf("%w: efeito fora de um run gerenciado", ErrPolicyDenied)
	}
	return gate.request(ctx, preview)
}

func (g *effectGate) request(ctx context.Context, preview EffectPreview) error {
	decisao, err := g.policy.Decide(ctx, g.graph, preview)
	if err != nil {
		return err
	}
	switch decisao {
	case PolicyAllow:
		return nil
	case PolicyDeny:
		return fmt.Errorf("%w: %s em %s", ErrPolicyDenied, preview.Effect, preview.Target)
	case PolicyRequireApproval:
		return g.exigeAprovacao(ctx, preview)
	default:
		return fmt.Errorf("%w: decisão desconhecida %q", ErrPolicyDenied, string(decisao))
	}
}

// exigeAprovacao resolve o ciclo pedido -> pendente -> decidido -> expirado.
func (g *effectGate) exigeAprovacao(ctx context.Context, preview EffectPreview) error {
	hash := preview.Hash()
	ap, existe, err := g.approvals.Get(ctx, g.run, hash)
	if err != nil {
		return err
	}
	if !existe {
		pedido := Approval{Run: g.run, Hash: hash, Preview: preview}
		if err := g.approvals.Request(ctx, pedido); err != nil {
			return err
		}
		return fmt.Errorf("%w: %s", ErrApprovalPending, preview.Effect)
	}
	if !ap.Decided {
		return fmt.Errorf("%w: %s", ErrApprovalPending, preview.Effect)
	}
	if !ap.Granted {
		return fmt.Errorf("%w: aprovação negada para %s", ErrPolicyDenied, preview.Effect)
	}
	if g.clock.Now().After(ap.ExpiresAt) {
		return fmt.Errorf("%w: aprovação de %s expirou", ErrApprovalExpired, preview.Effect)
	}
	return nil
}
