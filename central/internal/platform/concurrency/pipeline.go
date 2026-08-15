// Package concurrency implementa os helpers de pipeline limitado descritos em
// docs/architecture/mtga-go-concurrency.puml e na decisão 6 do OpenSpec
// add-central-go-platform.
//
// Regras de propriedade:
//   - todo estágio roda dentro de um errgroup com contexto compartilhado;
//     o primeiro erro cancela o contexto e para todos os produtores;
//   - todo envio seleciona também ctx.Done() — nenhum sender fica bloqueado
//     após o cancelamento;
//   - apenas o dono do canal (o estágio que o criou) o fecha, e somente após
//     todas as suas goroutines terminarem;
//   - toda capacidade de canal e todo número de workers vem de configuração,
//     nunca de decisão do código de runtime.
//
// Este é o único pacote do módulo autorizado a criar goroutines e canais;
// o teste de arquitetura em internal/arch garante a restrição.
package concurrency

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Send bloqueia até entregar v ou até o contexto ser cancelado.
func Send[T any](ctx context.Context, out chan<- T, v T) error {
	select {
	case out <- v:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Source publica os itens em um canal com a capacidade dada. É o dono do
// canal devolvido: fecha-o ao terminar ou ao ser cancelado.
func Source[T any](ctx context.Context, g *errgroup.Group, items []T, buffer int) <-chan T {
	out := make(chan T, buffer)
	g.Go(func() error {
		defer close(out)
		for _, v := range items {
			if err := Send(ctx, out, v); err != nil {
				return err
			}
		}
		return nil
	})
	return out
}
