package sqlitestore

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// TestDoisWorkersDisputamRecuperacaoSemDuplicar prova que dois workers
// sobre o MESMO banco recuperam um run interrompido exatamente uma vez:
// o lease decide e o perdedor recebe ErrLeaseHeld (tarefa 3.7).
func TestDoisWorkersDisputamRecuperacaoSemDuplicar(t *testing.T) {
	store := abre(t)
	interrompido(t, store, "run-disputado")
	engineA, _ := engineSobre(t, store, store, "wa")
	engineB, _ := engineSobre(t, store, store, "wb")

	var wg sync.WaitGroup
	resultados := make([]error, 2)
	for i, engine := range []*wf.Engine{engineA, engineB} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, resultados[i] = engine.Recover(context.Background(),
				"run-disputado", []string{"worker-a", "worker-b"}[i], time.Minute)
		}()
	}
	wg.Wait()

	vitorias, derrotas := 0, 0
	for _, err := range resultados {
		switch {
		case err == nil:
			vitorias++
		case errors.Is(err, wf.ErrLeaseHeld):
			derrotas++
		default:
			t.Fatalf("erro inesperado na disputa: %v", err)
		}
	}
	if vitorias != 1 || derrotas != 1 {
		t.Fatalf("exatamente um worker deveria vencer: %v", resultados)
	}
	final, err := store.Get(t.Context(), "run-disputado")
	if err != nil || final.Status != wf.RunSucceeded {
		t.Fatalf("run deveria concluir uma única vez: %+v (%v)", final, err)
	}
	steps, _ := store.Steps(t.Context(), "run-disputado")
	if len(steps) != 1 || steps[0].Node != "valida" {
		t.Fatalf("nó pendente deveria executar exatamente uma vez: %+v", steps)
	}
}

// storeQueMorre falha o Update após N confirmações (crash no checkpoint).
type storeQueMorre struct {
	*Store
	restantes int
}

func (s *storeQueMorre) Update(ctx context.Context, run wf.Run) error {
	if s.restantes <= 0 {
		return errors.New("crash simulado do processo")
	}
	s.restantes--
	return s.Store.Update(ctx, run)
}

// TestCrashERecuperacaoFimAFim mata o processo no primeiro checkpoint e
// prova que um novo worker retoma do nó confirmado e conclui.
func TestCrashERecuperacaoFimAFim(t *testing.T) {
	store := abre(t)
	moribundo, _ := engineSobre(t, &storeQueMorre{Store: store}, store, "w1")
	run, err := moribundo.Start(t.Context(), wf.Identity{Kind: "collection-sync", Version: 1}, nil)
	if err == nil {
		t.Fatal("primeiro checkpoint deveria falhar com o crash simulado")
	}
	sobrevivente, _ := engineSobre(t, store, store, "w2")
	recuperado, err := sobrevivente.Recover(t.Context(), run.ID, "worker-2", time.Minute)
	if err != nil || recuperado.Status != wf.RunSucceeded || recuperado.Outcome != "done" {
		t.Fatalf("recuperação fim a fim deveria concluir: %v / %+v", err, recuperado)
	}
}
