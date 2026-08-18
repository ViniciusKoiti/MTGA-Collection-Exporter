package collection

import (
	"fmt"
	"time"
)

// SourceKind identifica o tipo de fonte de coleção, em ordem de confiança
// (decisão 4 do design do companion).
type SourceKind string

// Fontes válidas.
const (
	SourceJSONImport   SourceKind = "json_import"   // export legado importado
	SourceDetailedLogs SourceKind = "detailed_logs" // logs do cliente MTGA
	SourceLegacyBridge SourceKind = "legacy_bridge" // scanner Python isolado
)

// ObservedQuantity é uma linha bruta observada: identidade crua da fonte,
// Arena ID quando reconhecido e quantidade declarada.
type ObservedQuantity struct {
	RawIdentity string
	Arena       ArenaID
	Quantity    int
}

// Observation é o que uma fonte devolve antes de qualquer normalização:
// identifica a fonte, o instante observado, as quantidades cruas e os
// diagnósticos de leitura.
type Observation struct {
	Source         SourceKind
	SourceInstance string
	ObservedAt     time.Time
	Quantities     []ObservedQuantity
	Diagnostics    []Diagnostic
}

// NewObservation valida os invariantes de uma observação: fonte conhecida,
// instante presente e quantidades não negativas.
func NewObservation(
	source SourceKind,
	instance string,
	observedAt time.Time,
	quantities []ObservedQuantity,
	diagnostics []Diagnostic,
) (Observation, error) {
	switch source {
	case SourceJSONImport, SourceDetailedLogs, SourceLegacyBridge:
	default:
		return Observation{}, fmt.Errorf("collection: fonte desconhecida %q", string(source))
	}
	if observedAt.IsZero() {
		return Observation{}, fmt.Errorf("collection: observação sem instante")
	}
	for _, q := range quantities {
		if q.Quantity < 0 {
			return Observation{}, fmt.Errorf("collection: quantidade negativa para %q", q.RawIdentity)
		}
	}
	return Observation{
		Source:         source,
		SourceInstance: instance,
		ObservedAt:     observedAt,
		Quantities:     append([]ObservedQuantity(nil), quantities...),
		Diagnostics:    append([]Diagnostic(nil), diagnostics...),
	}, nil
}
