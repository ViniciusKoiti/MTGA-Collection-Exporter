package metarecommend

import (
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Identity é a identidade pinada do grafo meta-deck-recommendation.
var Identity = wf.Identity{Kind: "meta-deck-recommendation", Version: 1}

// Outcomes terminais estáveis do grafo.
const (
	OutcomeDone        wf.OutcomeCode = "done"
	OutcomeSemSnapshot wf.OutcomeCode = "failed_no_snapshot"
	OutcomeSemCatalogo wf.OutcomeCode = "failed_stale_catalog"
)

// Definition compila o grafo com as dependências injetadas.
func Definition(deps Deps) wf.Definition {
	if deps.MaxLacunas == 0 {
		deps.MaxLacunas = 3
	}
	return wf.Definition{
		Identity: Identity,
		Initial:  "recomenda",
		Nodes: map[wf.NodeID]wf.Node{
			"recomenda": recomenda(deps),
		},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "recomenda", Outcome: "ok"}:           {Terminal: OutcomeDone},
			{From: "recomenda", Outcome: "sem_snapshot"}: {Terminal: OutcomeSemSnapshot, Falha: true},
			{From: "recomenda", Outcome: "sem_catalogo"}: {Terminal: OutcomeSemCatalogo, Falha: true},
		},
		Recovery: wf.RecoveryRestart, // recomendação é recomputável do zero
	}
}
