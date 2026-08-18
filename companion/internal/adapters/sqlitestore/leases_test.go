package sqlitestore

import (
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func runAtivo(t *testing.T, store *Store, id wf.RunID, atualizadoEm time.Time) {
	t.Helper()
	run := wf.Run{ID: id, Graph: wf.Identity{Kind: "collection-sync", Version: 1},
		Status: wf.RunActive, Current: "coleta", Version: 1,
		StartedAt: atualizadoEm, UpdatedAt: atualizadoEm}
	if err := store.Create(t.Context(), run); err != nil {
		t.Fatalf("create %s: %v", id, err)
	}
}

func TestLeaseDisputaEExpiracao(t *testing.T) {
	store := abre(t)
	base := time.Unix(1_700_000_000, 0).UTC()
	runAtivo(t, store, "run-1", base)
	ctx := t.Context()

	if ok, err := store.AcquireLease(ctx, "run-1", "w1", base, base.Add(time.Minute)); !ok || err != nil {
		t.Fatalf("primeiro lease: %v", err)
	}
	if ok, _ := store.AcquireLease(ctx, "run-1", "w2", base, base.Add(time.Minute)); ok {
		t.Fatal("lease vigente de outro dono não pode ser tomado")
	}
	if ok, err := store.AcquireLease(ctx, "run-1", "w1", base, base.Add(2*time.Minute)); !ok || err != nil {
		t.Fatalf("mesmo dono renova: %v", err)
	}
	depois := base.Add(3 * time.Minute)
	if ok, err := store.AcquireLease(ctx, "run-1", "w2", depois, depois.Add(time.Minute)); !ok || err != nil {
		t.Fatalf("lease expirado deveria ser tomado: %v", err)
	}
	if _, err := store.AcquireLease(ctx, "run-x", "w1", base, base.Add(time.Minute)); err == nil {
		t.Fatal("run inexistente deveria falhar")
	}
}

func TestRecoverableRunsEAbandono(t *testing.T) {
	store := abre(t)
	base := time.Unix(1_700_000_000, 0).UTC()
	runAtivo(t, store, "run-livre", base)
	runAtivo(t, store, "run-possuido", base)
	ctx := t.Context()
	if ok, err := store.AcquireLease(ctx, "run-possuido", "w1", base, base.Add(time.Hour)); !ok || err != nil {
		t.Fatalf("lease: %v", err)
	}

	recuperaveis, err := store.RecoverableRuns(ctx, base.Add(time.Minute))
	if err != nil || len(recuperaveis) != 1 || recuperaveis[0] != "run-livre" {
		t.Fatalf("apenas o run sem lease vigente é recuperável: %+v (%v)", recuperaveis, err)
	}

	corte := base.Add(30 * time.Minute)
	encerrados, err := store.AbandonExpired(ctx, corte, corte)
	if err != nil || encerrados != 1 {
		t.Fatalf("apenas o run sem lease deveria ser abandonado: %d (%v)", encerrados, err)
	}
	abandonado, _ := store.Get(ctx, "run-livre")
	if abandonado.Status != wf.RunFailed || abandonado.Outcome != wf.OutcomeAbandoned {
		t.Fatalf("desfecho de abandono incorreto: %+v", abandonado)
	}
	protegido, _ := store.Get(ctx, "run-possuido")
	if protegido.Status != wf.RunActive {
		t.Fatalf("run com lease vigente não pode ser abandonado: %+v", protegido)
	}
}
