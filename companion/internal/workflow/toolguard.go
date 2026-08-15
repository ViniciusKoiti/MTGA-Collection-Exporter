package workflow

import (
	"context"
	"fmt"
	"sync"
)

// ToolGuard aplica os orçamentos de chamadas de ferramenta de um run:
// total máximo e repetição máxima da mesma ferramenta com os mesmos
// argumentos (decisão 4 do design). Nós obtêm o guard do contexto e
// pedem autorização antes de cada chamada.
type ToolGuard struct {
	mu        sync.Mutex
	limites   Limits
	total     int
	repetidas map[string]int
}

func newToolGuard(l Limits) *ToolGuard {
	return &ToolGuard{limites: l, repetidas: make(map[string]int)}
}

// Authorize registra a intenção de chamar `tool` com o hash de argumentos
// dado e devolve ErrLimitExceeded quando qualquer orçamento estourar.
func (g *ToolGuard) Authorize(tool, argsHash string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.total+1 > g.limites.MaxToolCalls {
		return fmt.Errorf("%w: total de chamadas de ferramenta", ErrLimitExceeded)
	}
	chave := tool + "\x00" + argsHash
	if g.repetidas[chave]+1 > g.limites.MaxRepeatedCalls {
		return fmt.Errorf("%w: chamada repetida de %s", ErrLimitExceeded, tool)
	}
	g.total++
	g.repetidas[chave]++
	return nil
}

type toolGuardKey struct{}

// withToolGuard injeta o guard do run no contexto visto pelos nós.
func withToolGuard(ctx context.Context, g *ToolGuard) context.Context {
	return context.WithValue(ctx, toolGuardKey{}, g)
}

// GuardFromContext devolve o guard do run corrente; nós que chamam
// ferramentas DEVEM usá-lo — o harness tem asserção de comportamento
// proibido para chamadas sem autorização.
func GuardFromContext(ctx context.Context) (*ToolGuard, bool) {
	g, ok := ctx.Value(toolGuardKey{}).(*ToolGuard)
	return g, ok
}
