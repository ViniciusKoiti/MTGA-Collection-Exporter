package testkit

import (
	"fmt"
	"sort"
	"strings"
)

// DiagnosticBundle produz o pacote de diagnóstico redigido do resultado
// (tarefa 6.7): desfecho, journal de steps e trilha de eventos validados —
// com atributos em ordem estável. Estado, fixtures, alvos de efeito,
// caminhos e credenciais não existem neste formato por construção; o
// teste de golden de privacidade congela exatamente o que sai daqui.
func (r Result) DiagnosticBundle() string {
	var b strings.Builder
	fmt.Fprintf(&b, "bundle|v1|%s\n", r.Run.Graph)
	b.WriteString(r.Evidence())
	for _, ev := range r.Events {
		fmt.Fprintf(&b, "event|%02d|%s|%s|%dms|%s\n",
			ev.Step, ev.Outcome, ev.Causation, ev.DurationMS, attrsEstaveis(ev.Attrs))
	}
	return b.String()
}

// attrsEstaveis serializa os atributos permitidos em ordem alfabética.
func attrsEstaveis(attrs map[string]string) string {
	if len(attrs) == 0 {
		return "-"
	}
	chaves := make([]string, 0, len(attrs))
	for chave := range attrs {
		chaves = append(chaves, chave)
	}
	sort.Strings(chaves)
	pares := make([]string, 0, len(chaves))
	for _, chave := range chaves {
		pares = append(pares, chave+"="+attrs[chave])
	}
	return strings.Join(pares, ",")
}
