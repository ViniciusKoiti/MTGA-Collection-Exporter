package workflow_test

import (
	"context"
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func TestReplayDryRunNaoTocaProducaoNemDespachaEfeitos(t *testing.T) {
	exporta := noFn{"coleta", func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		preview := wf.EffectPreview{Effect: "export-write", Target: "decks/a.txt", PayloadHash: "h1"}
		if err := wf.RequestEffect(ctx, preview); err != nil {
			return st, "", err
		}
		return st, "ok", nil
	}}
	engine, _, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": exporta, "valida": passa("valida", "ok"),
	}, wf.Limits{})
	engine.WithPolicy(politicaFixa{wf.PolicyAllow}, memory.NewApprovals())

	original, err := engine.Start(t.Context(), identidade(), nil)
	if err != nil {
		t.Fatalf("run original: %v", err)
	}
	stepsAntes, _ := store.Steps(t.Context(), original.ID)

	scratch := memory.NewRunStore()
	replay, err := engine.ReplayDryRun(t.Context(), original.ID, nil, scratch)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if replay.Run.Status != wf.RunSucceeded || replay.Run.ID != wf.RunID("replay-"+string(original.ID)) {
		t.Fatalf("replay deveria concluir com ID derivado: %+v", replay.Run)
	}
	if len(replay.Efeitos) != 1 || replay.Efeitos[0].Effect != "export-write" {
		t.Fatalf("replay deveria colecionar os efeitos sem despachar: %+v", replay.Efeitos)
	}
	if len(replay.Steps) != 2 {
		t.Fatalf("journal do replay deveria viver no store descartável: %+v", replay.Steps)
	}
	stepsDepois, _ := store.Steps(t.Context(), original.ID)
	if len(stepsDepois) != len(stepsAntes) {
		t.Fatal("replay não pode acrescentar evidência ao run de produção")
	}
	if _, err := store.Get(t.Context(), replay.Run.ID); err == nil {
		t.Fatal("checkpoint do replay não pode entrar no store de produção")
	}
}

func TestReplayExigeRunTerminal(t *testing.T) {
	engine, _, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	ativo := runInterrompido(t, store, 1)
	if _, err := engine.ReplayDryRun(t.Context(), ativo.ID, nil, memory.NewRunStore()); err == nil {
		t.Fatal("replay de run não terminal deveria falhar")
	}
}
