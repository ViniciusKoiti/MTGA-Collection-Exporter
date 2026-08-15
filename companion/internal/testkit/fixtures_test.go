package testkit

import (
	"context"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

type noFn struct {
	id wf.NodeID
	fn func(context.Context, wf.State) (wf.State, wf.OutcomeCode, error)
}

func (n noFn) ID() wf.NodeID { return n.id }
func (n noFn) Execute(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
	return n.fn(ctx, st)
}

type aprovaTudo struct{}

func (aprovaTudo) Decide(context.Context, wf.Identity, wf.EffectPreview) (wf.PolicyDecision, error) {
	return wf.PolicyRequireApproval, nil
}

// cenarioExport monta o grafo de produção com um efeito que exige aprovação.
func cenarioExport(conceder bool) Scenario {
	registrar := func(r *wf.Registry) error {
		exporta := noFn{"exporta", func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
			preview := wf.EffectPreview{Effect: "export-write", Target: "decks/a.txt", PayloadHash: "h1"}
			if err := wf.RequestEffect(ctx, preview); err != nil {
				return st, "", err
			}
			return st, "ok", nil
		}}
		valida := noFn{"valida", func(_ context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
			return st, "ok", nil
		}}
		return r.Register(wf.Definition{
			Identity: wf.Identity{Kind: "approved-export", Version: 1},
			Initial:  "valida",
			Nodes:    map[wf.NodeID]wf.Node{"valida": valida, "exporta": exporta},
			Transitions: map[wf.TransitionKey]wf.Target{
				{From: "valida", Outcome: "ok"}:  {Next: "exporta"},
				{From: "exporta", Outcome: "ok"}: {Terminal: "done"},
			},
			Recovery: wf.RecoveryResume,
		})
	}
	return Scenario{
		Nome:       "approved-export-feliz",
		Graph:      wf.Identity{Kind: "approved-export", Version: 1},
		Registrar:  registrar,
		Estado:     map[string]int{"cartas": 3},
		Policy:     aprovaTudo{},
		Aprovacoes: map[string]Decisao{"export-write": {Conceder: conceder, Validade: time.Hour}},
		Inicio:     time.Unix(1_700_000_000, 0),
	}
}
