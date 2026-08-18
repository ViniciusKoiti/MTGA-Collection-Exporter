package workflow

import "errors"

// Erros tipados do runtime; consumidores comparam com errors.Is e nunca
// dependem do texto da mensagem.
var (
	// ErrInvalidDefinition indica definição de grafo rejeitada na validação.
	ErrInvalidDefinition = errors.New("workflow: definição de grafo inválida")

	// ErrDuplicateIdentity indica tentativa de registrar identidade já existente.
	ErrDuplicateIdentity = errors.New("workflow: identidade de grafo duplicada")

	// ErrUnknownIdentity indica busca por grafo não registrado.
	ErrUnknownIdentity = errors.New("workflow: grafo não registrado")

	// ErrLimitExceeded indica que um limite de execução foi atingido; o run
	// termina com resultado estável, nunca com pânico ou loop.
	ErrLimitExceeded = errors.New("workflow: limite de execução excedido")

	// ErrUnmappedOutcome indica outcome devolvido por um nó sem transição
	// compilada correspondente; é falha de execução, não decisão dinâmica.
	ErrUnmappedOutcome = errors.New("workflow: outcome sem transição compilada")
)
