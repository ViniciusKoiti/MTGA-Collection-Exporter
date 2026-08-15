package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// politicaFixa devolve sempre a mesma decisão.
type politicaFixa struct{ decisao wf.PolicyDecision }

func (p politicaFixa) Decide(context.Context, wf.Identity, wf.EffectPreview) (wf.PolicyDecision, error) {
	return p.decisao, nil
}

func preview() wf.EffectPreview {
	return wf.EffectPreview{Effect: "export-write", Target: "decks/a.txt", PayloadHash: "abc"}
}

// exportador pede o efeito exato antes de concluir.
func exportador() wf.Node {
	return noFn{"coleta", func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		if err := wf.RequestEffect(ctx, preview()); err != nil {
			return st, "", err
		}
		return st, "ok", nil
	}}
}

func ambientePolicy(t *testing.T, decisao wf.PolicyDecision) (*wf.Engine, *memory.Clock, *memory.Approvals) {
	t.Helper()
	engine, clock, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": exportador(), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	aprovacoes := memory.NewApprovals()
	engine.WithPolicy(politicaFixa{decisao}, aprovacoes)
	return engine, clock, aprovacoes
}

func TestPolicyDenyFalhaComOutcomeEstavel(t *testing.T) {
	engine, _, _ := ambientePolicy(t, wf.PolicyDeny)
	run, err := engine.Start(t.Context(), identidade(), nil)
	if !errors.Is(err, wf.ErrPolicyDenied) || run.Outcome != wf.OutcomePolicyDenied {
		t.Fatalf("esperava negação de política, veio: %v / %+v", err, run)
	}
}

func TestAprovacaoPausaERetomaComSeguranca(t *testing.T) {
	engine, clock, aprovacoes := ambientePolicy(t, wf.PolicyRequireApproval)
	run, err := engine.Start(t.Context(), identidade(), nil)
	if !errors.Is(err, wf.ErrApprovalPending) || run.Status != wf.RunWaiting {
		t.Fatalf("run deveria pausar aguardando aprovação: %v / %+v", err, run)
	}
	ap, existe, _ := aprovacoes.Get(t.Context(), run.ID, preview().Hash())
	if !existe || ap.Decided {
		t.Fatalf("solicitação pendente deveria existir: %+v", ap)
	}
	prazo := clock.Now().Add(time.Hour)
	if err := aprovacoes.Decide(t.Context(), run.ID, preview().Hash(), true, prazo); err != nil {
		t.Fatalf("decide: %v", err)
	}
	retomado, err := engine.Resume(t.Context(), run.ID)
	if err != nil || retomado.Status != wf.RunSucceeded {
		t.Fatalf("retomada aprovada deveria concluir: %v / %+v", err, retomado)
	}
}

func TestAprovacaoNegadaEncerraRun(t *testing.T) {
	engine, _, aprovacoes := ambientePolicy(t, wf.PolicyRequireApproval)
	run, _ := engine.Start(t.Context(), identidade(), nil)
	_ = aprovacoes.Decide(t.Context(), run.ID, preview().Hash(), false, time.Time{})
	retomado, err := engine.Resume(t.Context(), run.ID)
	if !errors.Is(err, wf.ErrPolicyDenied) || retomado.Outcome != wf.OutcomePolicyDenied {
		t.Fatalf("negação deveria encerrar o run: %v / %+v", err, retomado)
	}
}

func TestAprovacaoExpiradaEncerraRun(t *testing.T) {
	engine, clock, aprovacoes := ambientePolicy(t, wf.PolicyRequireApproval)
	run, _ := engine.Start(t.Context(), identidade(), nil)
	prazo := clock.Now().Add(time.Minute)
	_ = aprovacoes.Decide(t.Context(), run.ID, preview().Hash(), true, prazo)
	clock.Advance(2 * time.Minute)
	retomado, err := engine.Resume(t.Context(), run.ID)
	if !errors.Is(err, wf.ErrApprovalExpired) || retomado.Outcome != wf.OutcomeApprovalExpired {
		t.Fatalf("aprovação vencida deveria expirar o run: %v / %+v", err, retomado)
	}
}
