package sqlitestore

import (
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func evento(step int, outcome wf.OutcomeCode) wf.Event {
	return wf.Event{
		Schema: wf.EventSchema, Run: "run-000001", Step: step,
		Graph:       wf.Identity{Kind: "collection-sync", Version: 1},
		Correlation: "run-000001",
		At:          time.Unix(1_700_000_000, 0).UTC(),
		Outcome:     outcome, DurationMS: 5,
		Attrs: map[string]string{"status": "succeeded"},
	}
}

func TestJournalDeEventosComTimelinePaginada(t *testing.T) {
	store := abre(t)
	ctx := t.Context()
	for i, outcome := range []wf.OutcomeCode{"ok", "ok", "done"} {
		if err := store.Emit(ctx, evento(i+1, outcome)); err != nil {
			t.Fatalf("emit %d: %v", i, err)
		}
	}
	if err := store.Emit(ctx, func() wf.Event {
		alheio := evento(1, "ok")
		alheio.Run = "run-000002"
		return alheio
	}()); err != nil {
		t.Fatalf("emit alheio: %v", err)
	}

	pagina1, cursor, err := store.Timeline(ctx, "run-000001", 0, 2)
	if err != nil || len(pagina1) != 2 || cursor == 0 {
		t.Fatalf("página 1 inesperada: %d eventos, cursor %d (%v)", len(pagina1), cursor, err)
	}
	if pagina1[0].Step != 1 || pagina1[1].Step != 2 || pagina1[0].Attrs["status"] != "succeeded" {
		t.Fatalf("ordem/attrs do journal incorretos: %+v", pagina1)
	}
	pagina2, cursor, err := store.Timeline(ctx, "run-000001", cursor, 2)
	if err != nil || len(pagina2) != 1 || cursor != 0 {
		t.Fatalf("página 2 deveria fechar a timeline: %d eventos, cursor %d (%v)", len(pagina2), cursor, err)
	}
	if pagina2[0].Outcome != "done" || pagina2[0].Run != "run-000001" {
		t.Fatalf("evento final incorreto: %+v", pagina2[0])
	}
}
