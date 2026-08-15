package testkit

import (
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func TestCrashNoCheckpointPreservaUltimoEstadoConfirmado(t *testing.T) {
	sc := cenarioComFalha(func(n wf.Node, _ *memory.Clock) wf.Node { return n })
	sc.DecorarStore = func(s wf.RunStore) wf.RunStore { return CrashAposCheckpoints(s, 0) }
	h, err := New(sc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	res, err := h.Run(t.Context())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Err == nil {
		t.Fatal("crash simulado deveria propagar erro")
	}
	confirmado, err := h.Store.Get(t.Context(), res.Run.ID)
	if err != nil || confirmado.Version != 1 || confirmado.Current != "coleta" {
		t.Fatalf("último checkpoint confirmado deveria permanecer intacto: %+v (%v)", confirmado, err)
	}
}
