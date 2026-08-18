package testkit

import (
	"fmt"
	"strings"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Result reúne o desfecho devolvido pelo engine, o checkpoint persistido,
// o journal de steps, a trilha de eventos validados, os efeitos pendentes
// no outbox e o erro tipado (nil em sucesso).
type Result struct {
	Run        wf.Run
	Persistido wf.Run
	Steps      []wf.Step
	Events     []wf.Event
	Pendentes  []wf.EffectRecord
	Err        error
}

// Evidence produz a evidência normalizada e determinística do run: sequência
// ordenada de steps (nó, outcome, código estável de erro, duração pelo
// relógio controlado) e desfecho final. Cenário e seed iguais DEVEM produzir
// evidência idêntica (tarefa 5.8); payloads nunca aparecem aqui.
func (r Result) Evidence() string {
	var b strings.Builder
	for _, s := range r.Steps {
		fmt.Fprintf(&b, "step|%02d|%s|%s|%s|%dms\n",
			s.Index, s.Node, s.Outcome, s.Err, s.Finished.Sub(s.Started).Milliseconds())
	}
	fmt.Fprintf(&b, "run|%s|%s|%s|v%d\n", r.Run.Graph, r.Run.Status, r.Run.Outcome, r.Run.Version)
	return b.String()
}
