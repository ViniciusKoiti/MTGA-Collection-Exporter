// Package telemetryflush implementa o grafo telemetry-flush v1 (tarefa 4.4
// do OpenSpec add-graph-workflow-harness): verificação de consentimento,
// minimização por allowlist, lote limitado, envio com retry limitado,
// acknowledgement e fallback local-only — nada sai do dispositivo sem
// opt-in explícito e eventos não confirmados permanecem locais.
package telemetryflush

import (
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Deps são os ports e orçamentos do grafo; tudo vem da composição.
type Deps struct {
	Consent ports.ConsentStore
	Queue   ports.TelemetryQueue
	Sink    ports.TelemetrySink
	// AtributosPermitidos é a allowlist de chaves; valores acima de
	// MaxValorBytes derrubam o atributo, nunca o evento inteiro.
	AtributosPermitidos map[string]bool
	MaxValorBytes       int
	TamanhoLote         int
	MaxTentativas       int
}

// Estado é o estado tipado do run.
type Estado struct {
	Lote        []ports.TelemetryEvent
	Minimizados int // atributos descartados pela minimização
	UltimoSeq   int
	Tentativas  int
	Enviados    int
}

// estadoDe normaliza o estado recebido pelo nó.
func estadoDe(st any) Estado {
	estado, _ := st.(Estado)
	return estado
}
