package telemetryflush

import (
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Identity é a identidade pinada do grafo telemetry-flush.
var Identity = wf.Identity{Kind: "telemetry-flush", Version: 1}

// Outcomes terminais estáveis. Local-only é SUCESSO: manter os eventos no
// dispositivo é o comportamento correto sem opt-in ou sem central.
const (
	OutcomeDone      wf.OutcomeCode = "done"
	OutcomeLocalOnly wf.OutcomeCode = "done_local_only"
	OutcomeSemFila   wf.OutcomeCode = "done_nothing_pending"
)

// Definition compila o grafo com as dependências injetadas.
func Definition(deps Deps) wf.Definition {
	return wf.Definition{
		Identity: Identity,
		Initial:  "consente",
		Nodes: map[wf.NodeID]wf.Node{
			"consente": consente(deps),
			"minimiza": minimiza(deps),
			"envia":    envia(deps),
			"confirma": confirma(deps),
		},
		Transitions: map[wf.TransitionKey]wf.Target{
			{From: "consente", Outcome: "ok"}:          {Next: "minimiza"},
			{From: "consente", Outcome: "sem_optin"}:   {Terminal: OutcomeLocalOnly},
			{From: "minimiza", Outcome: "ok"}:          {Next: "envia"},
			{From: "minimiza", Outcome: "fila_vazia"}:  {Terminal: OutcomeSemFila},
			{From: "envia", Outcome: "ok"}:             {Next: "confirma"},
			{From: "envia", Outcome: "envio_esgotado"}: {Terminal: OutcomeLocalOnly},
			{From: "confirma", Outcome: "ok"}:          {Terminal: OutcomeDone},
		},
		Recovery: wf.RecoveryResume,
	}
}
