// Package workflow define os contratos do runtime de grafos do companion
// (OpenSpec add-graph-workflow-harness, decisões 1-4 do design):
// toda atividade multi-etapas é um grafo tipado, versionado e limitado,
// compilado em código; nenhuma definição vem de modelo ou catálogo remoto.
package workflow

import (
	"fmt"
	"time"
)

// Kind identifica a atividade (collection-sync, meta-deck-recommendation,
// approved-export, telemetry-flush, development-scenario).
type Kind string

// Identity é a identidade única de um grafo: tipo + versão compilada.
// Um run fixa a identidade e nunca migra de versão no meio da execução.
type Identity struct {
	Kind    Kind
	Version int
}

func (id Identity) String() string {
	return fmt.Sprintf("%s/v%d", id.Kind, id.Version)
}

// NodeID nomeia um nó dentro de um grafo.
type NodeID string

// OutcomeCode é o resultado tipado devolvido por um nó; as transições são
// resolvidas exclusivamente a partir dele — nunca de texto livre.
type OutcomeCode string

// RecoveryPolicy define o comportamento após interrupção.
type RecoveryPolicy string

// Políticas de recuperação válidas.
const (
	RecoveryResume  RecoveryPolicy = "resume"  // retoma do último checkpoint
	RecoveryRestart RecoveryPolicy = "restart" // reinicia do nó inicial
	RecoveryFail    RecoveryPolicy = "fail"    // encerra com falha terminal
)

func (p RecoveryPolicy) valida() error {
	switch p {
	case RecoveryResume, RecoveryRestart, RecoveryFail:
		return nil
	}
	return fmt.Errorf("%w: recovery policy %q", ErrInvalidDefinition, string(p))
}

// Limits são os limites de execução de um grafo. Defaults da decisão 4;
// uma definição pode reduzir os valores mas nunca exceder os tetos duros
// configurados no Registry.
type Limits struct {
	MaxSteps         int
	MaxToolCalls     int
	MaxRepeatedCalls int
	ActiveDeadline   time.Duration
}

// DefaultLimits são aplicados quando a definição deixa campos zerados.
func DefaultLimits() Limits {
	return Limits{
		MaxSteps:         8,
		MaxToolCalls:     6,
		MaxRepeatedCalls: 2,
		ActiveDeadline:   30 * time.Second,
	}
}
