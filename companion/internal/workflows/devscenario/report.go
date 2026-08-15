package devscenario

import (
	"context"
	"fmt"
	"strings"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// avalia compara a evidência com a asserção declarada no cenário e produz
// o relatório redigido: nome, grafo alvo, desfecho, veredicto e evidência
// normalizada — nunca fixtures, payloads ou caminhos.
func avalia() wf.Node {
	return no{"avalia", func(_ context.Context, estado Estado) (Estado, wf.OutcomeCode, error) {
		veredicto := "aprovado"
		detalhe := "sem asserção declarada"
		if esperado := estado.Parsed.Esperado; esperado != nil {
			quer := fmt.Sprintf("%s/%s", esperado.Status, esperado.Outcome)
			if estado.Desfecho != quer {
				veredicto = "reprovado"
				detalhe = fmt.Sprintf("esperava %s, veio %s", quer, estado.Desfecho)
			} else {
				detalhe = "desfecho conferido: " + quer
			}
		}
		var b strings.Builder
		fmt.Fprintf(&b, "relatorio|scenario/v1|%s\n", estado.Parsed.Nome)
		fmt.Fprintf(&b, "alvo|%s/v%d\n", estado.Parsed.Graph.Kind, estado.Parsed.Graph.Version)
		fmt.Fprintf(&b, "desfecho|%s\n", estado.Desfecho)
		fmt.Fprintf(&b, "veredicto|%s|%s\n", veredicto, detalhe)
		b.WriteString(estado.Evidencia)
		estado.Relatorio = b.String()
		if veredicto == "reprovado" {
			return estado, "reprovado", nil
		}
		return estado, "aprovado", nil
	}}
}
