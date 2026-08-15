package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrEventoProibido indica evento rejeitado no sink por conter campo,
// chave ou valor fora do contrato de privacidade (tarefa 6.3).
var ErrEventoProibido = errors.New("workflow: evento com campo proibido")

// fragmentosProibidos derrubam qualquer valor de atributo que se pareça
// com caminho, credencial, prompt ou payload estruturado.
var fragmentosProibidos = []string{
	"\\", "/", "{", "bearer ", "password", "secret", "token", "prompt",
}

// ValidatingSink decora um EventSink rejeitando eventos proibidos antes de
// qualquer persistência; é a última linha de defesa de privacidade e por
// isso falha alto em vez de redigir silenciosamente.
type ValidatingSink struct {
	Next EventSink
}

// Emit valida e repassa; rejeição nunca chega ao sink decorado.
func (v ValidatingSink) Emit(ctx context.Context, ev Event) error {
	if err := validaEvento(ev); err != nil {
		return err
	}
	return v.Next.Emit(ctx, ev)
}

func validaEvento(ev Event) error {
	if ev.Schema != EventSchema {
		return fmt.Errorf("%w: schema %q", ErrEventoProibido, ev.Schema)
	}
	if ev.Run == "" || ev.Graph.Kind == "" {
		return fmt.Errorf("%w: evento sem identificação de run/grafo", ErrEventoProibido)
	}
	for chave, valor := range ev.Attrs {
		if !AttrAllowlist[chave] {
			return fmt.Errorf("%w: chave %q fora da allowlist", ErrEventoProibido, chave)
		}
		if len(valor) > MaxAttrValueLen {
			return fmt.Errorf("%w: valor de %q excede %d bytes", ErrEventoProibido,
				chave, MaxAttrValueLen)
		}
		minusculo := strings.ToLower(valor)
		for _, fragmento := range fragmentosProibidos {
			if strings.Contains(minusculo, fragmento) {
				return fmt.Errorf("%w: valor de %q contém fragmento vetado", ErrEventoProibido, chave)
			}
		}
	}
	return nil
}
