package storetest

import (
	"errors"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// versaoOtimista cobre o contrato central do checkpoint: só a versão
// armazenada + 1 é aceita; qualquer outra devolve ErrVersionConflict e
// não altera o estado persistido.
func versaoOtimista(t *testing.T, store wf.RunStore) {
	run := runExemplo()
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
	run.Version = 2
	run.Current = "valida"
	if err := store.Update(t.Context(), run); err != nil {
		t.Fatalf("update sequencial deveria passar: %v", err)
	}
	repetido := run // versão 2 de novo: escritor concorrente atrasado
	repetido.Current = "intruso"
	if err := store.Update(t.Context(), repetido); !errors.Is(err, wf.ErrVersionConflict) {
		t.Fatalf("esperava ErrVersionConflict, veio: %v", err)
	}
	saltado := run
	saltado.Version = 9 // pulo de versão também é conflito
	if err := store.Update(t.Context(), saltado); !errors.Is(err, wf.ErrVersionConflict) {
		t.Fatalf("esperava ErrVersionConflict no salto, veio: %v", err)
	}
	lido, err := store.Get(t.Context(), run.ID)
	if err != nil || lido.Version != 2 || lido.Current != "valida" {
		t.Fatalf("conflitos não podem alterar o checkpoint: %+v (%v)", lido, err)
	}
}

// stepsOrdenados garante journal completo e em ordem de inserção, mesmo
// com índices repetidos entre sessões de retomada.
func stepsOrdenados(t *testing.T, store wf.RunStore) {
	run := runExemplo()
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create: %v", err)
	}
	base := time.Unix(1_700_000_000, 0).UTC()
	entradas := []wf.Step{
		{Run: run.ID, Index: 1, Node: "coleta", Attempt: 1, Outcome: "ok"},
		{Run: run.ID, Index: 2, Node: "valida", Attempt: 1, Err: "approval_pending"},
		{Run: run.ID, Index: 1, Node: "valida", Attempt: 1, Outcome: "ok"},
	}
	for i, step := range entradas {
		step.Started = base.Add(time.Duration(i) * time.Second)
		step.Finished = step.Started.Add(time.Second)
		if err := store.AppendStep(t.Context(), step); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	steps, err := store.Steps(t.Context(), run.ID)
	if err != nil {
		t.Fatalf("steps: %v", err)
	}
	if len(steps) != 3 || steps[0].Node != "coleta" ||
		steps[1].Err != "approval_pending" || steps[2].Index != 1 {
		t.Fatalf("journal fora de ordem ou incompleto: %+v", steps)
	}
}
