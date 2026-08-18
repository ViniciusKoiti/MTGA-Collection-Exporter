package approvedexport

import (
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Identity é a identidade pinada do grafo approved-export.
var Identity = wf.Identity{Kind: "approved-export", Version: 1}

// Outcomes terminais estáveis do grafo.
const (
	OutcomeDone        wf.OutcomeCode = "done"
	OutcomeSemSnapshot wf.OutcomeCode = "failed_no_snapshot"
)

// Definition compila o grafo com as dependências injetadas.
func Definition(deps Deps) wf.Definition {
	return wf.Definition{
		Identity: Identity,
		Initial:  "carrega",
		Nodes: map[wf.NodeID]wf.Node{
			"carrega": carrega(deps),
			"aprova":  aprova(),
			"escreve": escreve(deps),
		},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "carrega", Outcome: "ok"}:           {Next: "aprova"},
			{From: "carrega", Outcome: "sem_snapshot"}: {Terminal: OutcomeSemSnapshot, Falha: true},
			{From: "aprova", Outcome: "ok"}:            {Next: "escreve"},
			{From: "escreve", Outcome: "ok"}:           {Terminal: OutcomeDone},
		},
		Recovery: wf.RecoveryResume,
	}
}
