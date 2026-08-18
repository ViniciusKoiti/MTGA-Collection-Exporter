package collectionsync

import (
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Identity é a identidade pinada do grafo collection-sync.
var Identity = wf.Identity{Kind: "collection-sync", Version: 1}

// Outcomes terminais estáveis do grafo.
const (
	OutcomeDone     wf.OutcomeCode = "done"
	OutcomeSemFonte wf.OutcomeCode = "failed_no_source"
	OutcomeVazia    wf.OutcomeCode = "failed_empty_collection"
)

// Definition compila o grafo com as dependências injetadas. As transições
// são a única forma de avançar; nenhum nó decide destino por conta própria.
func Definition(deps Deps) wf.Definition {
	return wf.Definition{
		Identity: Identity,
		Initial:  "observa",
		Nodes: map[wf.NodeID]wf.Node{
			"observa":   observa(deps),
			"normaliza": normaliza(deps),
			"valida":    valida(),
			"persiste":  persiste(deps),
			"projeta":   projeta(deps),
		},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "observa", Outcome: "ok"}:        {Next: "normaliza"},
			{From: "observa", Outcome: "sem_fonte"}: {Terminal: OutcomeSemFonte, Falha: true},
			{From: "normaliza", Outcome: "ok"}:      {Next: "valida"},
			{From: "valida", Outcome: "ok"}:         {Next: "persiste"},
			{From: "valida", Outcome: "vazia"}:      {Terminal: OutcomeVazia, Falha: true},
			{From: "persiste", Outcome: "ok"}:       {Next: "projeta"},
			{From: "projeta", Outcome: "ok"}:        {Terminal: OutcomeDone},
		},
		Recovery: wf.RecoveryResume,
	}
}
