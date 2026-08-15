package workflow

import "context"

// State é o estado tipado do run, opaco para o engine; cada grafo define o
// seu tipo concreto e os nós devolvem o próximo estado imutavelmente.
type State any

// Node executa uma etapa e devolve o próximo estado e um outcome tipado.
// Operações puras de domínio (normalização, legalidade, ranking) são funções
// comuns chamadas de dentro de Execute, nunca grafos próprios.
type Node interface {
	ID() NodeID
	Execute(ctx context.Context, st State) (State, OutcomeCode, error)
}

// Target é o destino de uma transição: exatamente um de Next/Terminal é
// preenchido. Terminal encerra o run com o outcome dado; Falha marca o
// desfecho terminal como falha de negócio (status failed) em vez de
// sucesso — a distinção é declarada no grafo, nunca decidida pelo nó.
type Target struct {
	Next     NodeID
	Terminal OutcomeCode
	Falha    bool
}

// EhTerminal informa se o alvo encerra o run.
func (t Target) EhTerminal() bool { return t.Terminal != "" }

// TransitionKey indexa a tabela de transições compilada: nó de origem +
// outcome devolvido. Só o que está na tabela pode acontecer.
type TransitionKey struct {
	From    NodeID
	Outcome OutcomeCode
}

// Definition é um grafo compilado em código: identidade, nó inicial, nós,
// transições, limites e política de recuperação. Definições são valores
// imutáveis registrados na composição do processo.
type Definition struct {
	Identity    Identity
	Initial     NodeID
	Nodes       map[NodeID]Node
	Transitions map[TransitionKey]Target
	Limits      Limits
	Recovery    RecoveryPolicy
}

// limitesEfetivos preenche campos zerados com os defaults da decisão 4.
func (d Definition) limitesEfetivos() Limits {
	efetivo := d.Limits
	padrao := DefaultLimits()
	if efetivo.MaxSteps == 0 {
		efetivo.MaxSteps = padrao.MaxSteps
	}
	if efetivo.MaxToolCalls == 0 {
		efetivo.MaxToolCalls = padrao.MaxToolCalls
	}
	if efetivo.MaxRepeatedCalls == 0 {
		efetivo.MaxRepeatedCalls = padrao.MaxRepeatedCalls
	}
	if efetivo.MaxPayloadBytes == 0 {
		efetivo.MaxPayloadBytes = padrao.MaxPayloadBytes
	}
	if efetivo.ActiveDeadline == 0 {
		efetivo.ActiveDeadline = padrao.ActiveDeadline
	}
	return efetivo
}
