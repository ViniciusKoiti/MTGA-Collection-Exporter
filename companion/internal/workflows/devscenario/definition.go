package devscenario

import (
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Identity é a identidade pinada do grafo development-scenario.
var Identity = wf.Identity{Kind: "development-scenario", Version: 1}

// Outcomes terminais estáveis do grafo.
const (
	OutcomeDone      wf.OutcomeCode = "done"
	OutcomeInvalido  wf.OutcomeCode = "failed_invalid_scenario"
	OutcomeSemGrafo  wf.OutcomeCode = "failed_unknown_graph"
	OutcomeReprovado wf.OutcomeCode = "failed_assertion"
)

// Definition compila o grafo com o catálogo injetado.
func Definition(deps Deps) wf.Definition {
	return wf.Definition{
		Identity: Identity,
		Initial:  "prepara",
		Nodes: map[wf.NodeID]wf.Node{
			"prepara": prepara(),
			"executa": executa(deps),
			"avalia":  avalia(),
		},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "prepara", Outcome: "ok"}:        {Next: "executa"},
			{From: "prepara", Outcome: "invalido"}:  {Terminal: OutcomeInvalido, Falha: true},
			{From: "executa", Outcome: "ok"}:        {Next: "avalia"},
			{From: "executa", Outcome: "sem_grafo"}: {Terminal: OutcomeSemGrafo, Falha: true},
			{From: "avalia", Outcome: "aprovado"}:   {Terminal: OutcomeDone},
			{From: "avalia", Outcome: "reprovado"}:  {Terminal: OutcomeReprovado, Falha: true},
		},
		Recovery: wf.RecoveryRestart, // cenário é reexecutável do zero
	}
}
