package concurrency

import (
	"context"
	"sort"
)

// Reduce drena o canal em uma única goroutine (a do chamador) e devolve os
// itens em ordem determinística definida por less. O paralelismo termina
// aqui: a redução é single-owner para que a publicação permaneça auditável,
// como exige o redutor estável do diagrama de concorrência.
//
// O chamador deve garantir, via errgroup compartilhado, que os produtores
// fecharão `in` mesmo sob cancelamento; Reduce também observa ctx para não
// bloquear caso o dono do canal já tenha sido cancelado.
func Reduce[T any](ctx context.Context, in <-chan T, less func(a, b T) bool) ([]T, error) {
	var items []T
	for {
		select {
		case v, ok := <-in:
			if !ok {
				sort.SliceStable(items, func(i, j int) bool { return less(items[i], items[j]) })
				return items, nil
			}
			items = append(items, v)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Batches fatia os itens já reduzidos em lotes limitados para escrita.
// Escrita em lote limitado substitui "uma goroutine por registro".
func Batches[T any](items []T, size int) [][]T {
	if size < 1 || len(items) == 0 {
		return nil
	}
	batches := make([][]T, 0, (len(items)+size-1)/size)
	for start := 0; start < len(items); start += size {
		end := min(start+size, len(items))
		batches = append(batches, items[start:end])
	}
	return batches
}
