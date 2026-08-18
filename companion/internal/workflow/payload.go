package workflow

import (
	"encoding/json"
	"fmt"
)

// ErrPayloadExceeded indica estado serializado maior que o limite do grafo.
// Também satisfaz errors.Is(err, ErrLimitExceeded).
var ErrPayloadExceeded = fmt.Errorf("%w: payload do estado", ErrLimitExceeded)

// validaPayload garante que o estado do checkpoint é serializável e cabe no
// orçamento de bytes; estados gigantes indicariam coleção inteira dentro do
// run, o que a spec de observabilidade proíbe.
func validaPayload(st State, maxBytes int) error {
	serializado, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("workflow: estado não serializável para checkpoint: %w", err)
	}
	if len(serializado) > maxBytes {
		return fmt.Errorf("%w: %d > %d bytes", ErrPayloadExceeded, len(serializado), maxBytes)
	}
	return nil
}
