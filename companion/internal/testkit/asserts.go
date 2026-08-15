package testkit

import (
	"fmt"
	"strings"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Asserções da tarefa 5.5: desfecho, transições ordenadas, tentativas,
// efeitos, orçamento, banco, redação e comportamento proibido.

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

// AssertTentativasDoNo valida quantas vezes um nó foi tentado no journal.
func (r Result) AssertTentativasDoNo(no wf.NodeID, esperadas int) error {
	tentativas := 0
	for _, s := range r.Steps {
		if s.Node == no {
			tentativas++
		}
	}
	if tentativas != esperadas {
		return fmt.Errorf("testkit: nó %q tentado %d vezes, esperava %d", no, tentativas, esperadas)
	}
	return nil
}

// AssertEfeitosPendentes valida o outbox capturado ao fim do cenário.
func (r Result) AssertEfeitosPendentes(quantidade int) error {
	if len(r.Pendentes) != quantidade {
		return fmt.Errorf("testkit: %d efeitos pendentes, esperava %d", len(r.Pendentes), quantidade)
	}
	return nil
}

// AssertCheckpointConsistente compara o run devolvido com o persistido.
func (r Result) AssertCheckpointConsistente() error {
	if r.Persistido.Status != r.Run.Status || r.Persistido.Version != r.Run.Version ||
		r.Persistido.Outcome != r.Run.Outcome {
		return fmt.Errorf("testkit: banco divergente do desfecho: %+v vs %+v",
			r.Persistido, r.Run)
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

// AssertEventos valida a quantidade e o envelope dos eventos emitidos.
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

// AssertProibido garante que a evidência não contém fragmentos vetados.
func (r Result) AssertProibido(fragmentos ...string) error {
	evidencia := r.Evidence()
	for _, f := range fragmentos {
		if f != "" && strings.Contains(evidencia, f) {
			return fmt.Errorf("testkit: evidência contém fragmento proibido %q", f)
		}
	}
	return nil
}
