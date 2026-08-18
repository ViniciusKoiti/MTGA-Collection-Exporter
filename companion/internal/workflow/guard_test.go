package workflow_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func TestToolGuardLimitaTotalERepeticao(t *testing.T) {
	usaFerramenta := noFn{"coleta", func(ctx context.Context, st wf.State) (wf.State, wf.OutcomeCode, error) {
		guard, ok := wf.GuardFromContext(ctx)
		if !ok {
			t.Fatal("guard ausente do contexto do nó")
		}
		if err := guard.Authorize("scryfall", "a"); err != nil {
			return st, "", err
		}
		if err := guard.Authorize("scryfall", "a"); err != nil {
			return st, "", err // segunda repetição idêntica estoura MaxRepeatedCalls=1
		}
		return st, "ok", nil
	}}
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": usaFerramenta, "valida": passa("valida", "ok"),
	}, wf.Limits{MaxToolCalls: 6, MaxRepeatedCalls: 1})
	run, err := engine.Start(t.Context(), identidade(), nil)
	if !errors.Is(err, wf.ErrLimitExceeded) || run.Outcome != wf.OutcomeToolLimit {
		t.Fatalf("esperava limite de ferramenta, veio: %v / %+v", err, run)
	}
}

func TestRunsConcorrentesNaoInterferem(t *testing.T) {
	engine, _, store := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	var wg sync.WaitGroup
	resultados := make([]wf.Run, 16)
	for i := range resultados {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resultados[i], _ = engine.Start(t.Context(), identidade(), i)
		}()
	}
	wg.Wait()
	vistos := map[wf.RunID]bool{}
	for _, run := range resultados {
		if run.Status != wf.RunSucceeded || vistos[run.ID] {
			t.Fatalf("run concorrente inconsistente: %+v", run)
		}
		vistos[run.ID] = true
		persistido, err := store.Get(t.Context(), run.ID)
		if err != nil || persistido.Version != run.Version {
			t.Fatalf("checkpoint divergente: %+v (%v)", persistido, err)
		}
	}
}
