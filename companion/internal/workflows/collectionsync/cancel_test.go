package collectionsync

import (
	"context"
	"errors"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// TestRepetirMesmaObservacaoEhIdempotente: same observation at the same
// controlled instant reuses the committed snapshot (task 3.3).
func TestRepetirMesmaObservacaoEhIdempotente(t *testing.T) {
	sc, snapshots := ambiente(t, fonte([]collection.ObservedQuantity{
		{RawIdentity: "101", Arena: 101, Quantity: 4},
	}, nil))
	primeira := executa(t, sc)
	segunda := executa(t, sc) // mesmo cenário, mesma seed, mesmo store
	if segunda.Err != nil || segunda.Run.Status != wf.RunSucceeded {
		t.Fatalf("re-execução deveria ser idempotente: %v / %+v", segunda.Err, segunda.Run)
	}
	if primeira.Run.State.(Estado).Snapshot != segunda.Run.State.(Estado).Snapshot {
		t.Fatal("mesma observação deve reutilizar o mesmo snapshot")
	}
	if snap, existe, _ := snapshots.Latest(t.Context()); !existe || snap.TotalCartas() != 4 {
		t.Fatalf("snapshot único esperado: %+v", snap)
	}
}

// TestCancelamentoEncerraSemCommit cobre o caminho golden de cancelamento
// (tarefa 4.7): o run termina como cancelled com outcome estável e o
// snapshot autoritativo permanece intocado.
func TestCancelamentoEncerraSemCommit(t *testing.T) {
	sc, snapshots := ambiente(t, fonte([]collection.ObservedQuantity{
		{RawIdentity: "101", Arena: 101, Quantity: 4},
	}, nil))
	h, err := testkit.New(sc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancelamento chega antes do primeiro nó
	res, err := h.Run(ctx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !errors.Is(res.Err, context.Canceled) {
		t.Fatalf("erro tipado deveria ser cancelamento: %v", res.Err)
	}
	if res.Run.Status != wf.RunCancelled || res.Run.Outcome != wf.OutcomeCancelled {
		t.Fatalf("desfecho de cancelamento inesperado: %+v", res.Run)
	}
	if _, existe, _ := snapshots.Latest(ctx); existe {
		t.Fatal("cancelamento não pode commitar snapshot")
	}
}
