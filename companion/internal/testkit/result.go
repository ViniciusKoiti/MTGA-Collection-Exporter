package testkit

import (
	"fmt"
	"strings"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Result reúne o desfecho persistido, o journal de steps, a trilha de
// eventos validados e o erro tipado devolvido pelo engine (nil em sucesso).
type Result struct {
	Run    wf.Run
	Steps  []wf.Step
	Events []wf.Event
	Err    error
}

// AssertEventos valida a quantidade de eventos emitidos e o envelope de
// cada um (schema versionado + correlação do run).
func (r Result) AssertEventos(quantidade int) error {
	if len(r.Events) != quantidade {
		return fmt.Errorf("testkit: %d eventos, esperava %d", len(r.Events), quantidade)
	}
	for i, ev := range r.Events {
		if ev.Schema != wf.EventSchema || ev.Correlation != string(r.Run.ID) {
			return fmt.Errorf("testkit: evento %d fora do envelope: %+v", i, ev)
		}
	}
	return nil
}

// AssertOrcamento garante que a execução coube no orçamento de passos.
func (r Result) AssertOrcamento(maxSteps int) error {
	if len(r.Steps) > maxSteps {
		return fmt.Errorf("testkit: %d steps excedem o orçamento de %d", len(r.Steps), maxSteps)
	}
	return nil
}

// Evidence produz a evidência normalizada e determinística do run: sequência
// ordenada de steps (nó, outcome, erro estável, duração pelo relógio
// controlado) e desfecho final. Cenário e seed iguais DEVEM produzir
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

// AssertDesfecho valida status e outcome finais.
func (r Result) AssertDesfecho(status wf.RunStatus, outcome wf.OutcomeCode) error {
	if r.Run.Status != status || r.Run.Outcome != outcome {
		return fmt.Errorf("testkit: desfecho %s/%s, esperado %s/%s",
			r.Run.Status, r.Run.Outcome, status, outcome)
	}
	return nil
}

// AssertTransicoes valida a sequência ordenada de nós executados.
func (r Result) AssertTransicoes(nos ...wf.NodeID) error {
	if len(r.Steps) != len(nos) {
		return fmt.Errorf("testkit: %d steps, esperava %d", len(r.Steps), len(nos))
	}
	for i, esperado := range nos {
		if r.Steps[i].Node != esperado {
			return fmt.Errorf("testkit: step %d executou %q, esperava %q",
				i+1, r.Steps[i].Node, esperado)
		}
	}
	return nil
}

// AssertProibido garante que a evidência não contém nenhum dos fragmentos
// proibidos (caminhos, credenciais, payloads) — asserção de comportamento
// proibido da tarefa 5.5.
func (r Result) AssertProibido(fragmentos ...string) error {
	evidencia := r.Evidence()
	for _, f := range fragmentos {
		if f != "" && strings.Contains(evidencia, f) {
			return fmt.Errorf("testkit: evidência contém fragmento proibido %q", f)
		}
	}
	return nil
}
