package workflow_test

import (
	"errors"
	"strings"
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

func eventoValido() wf.Event {
	return wf.Event{
		Schema: wf.EventSchema, Run: "run-000001",
		Graph: wf.Identity{Kind: "collection-sync", Version: 1},
		Attrs: map[string]string{"status": "succeeded", "count": "42"},
	}
}

func TestSinkRejeitaCamposProibidos(t *testing.T) {
	casos := map[string]func(*wf.Event){
		"schema desconhecido": func(ev *wf.Event) { ev.Schema = "v0" },
		"sem identificação":   func(ev *wf.Event) { ev.Run = "" },
		"chave fora da allowlist": func(ev *wf.Event) {
			ev.Attrs["collection_json"] = "x"
		},
		"payload estruturado": func(ev *wf.Event) {
			ev.Attrs["provider"] = `{"grpId":123,"qty":4}`
		},
		"caminho windows": func(ev *wf.Event) {
			ev.Attrs["provider"] = `C:\Users\u\mtga_collection.json`
		},
		"caminho unix": func(ev *wf.Event) {
			ev.Attrs["provider"] = "home/u/logs/player.log"
		},
		"credencial": func(ev *wf.Event) {
			ev.Attrs["error_code"] = "Bearer abc123"
		},
		"segredo": func(ev *wf.Event) {
			ev.Attrs["error_code"] = "password=hunter2"
		},
		"prompt de modelo": func(ev *wf.Event) {
			ev.Attrs["status"] = "prompt: monte um deck"
		},
		"log gigante": func(ev *wf.Event) {
			ev.Attrs["status"] = strings.Repeat("linha de log ", 20)
		},
	}
	sink := wf.ValidatingSink{Next: memory.NewEvents()}
	for nome, mutacao := range casos {
		ev := eventoValido()
		mutacao(&ev)
		if err := sink.Emit(t.Context(), ev); !errors.Is(err, wf.ErrEventoProibido) {
			t.Errorf("%s: sink deveria rejeitar, veio %v", nome, err)
		}
	}
	if err := sink.Emit(t.Context(), eventoValido()); err != nil {
		t.Fatalf("evento válido deveria passar: %v", err)
	}
}

func TestEngineEmiteEventosValidadosPorStepEDesfecho(t *testing.T) {
	engine, _, _ := ambiente(t, map[wf.NodeID]wf.Node{
		"coleta": passa("coleta", "ok"), "valida": passa("valida", "ok"),
	}, wf.Limits{})
	trilha := memory.NewEvents()
	engine.WithEvents(trilha)
	run, err := engine.Start(t.Context(), identidade(), nil)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	eventos := trilha.Trilha()
	if len(eventos) != 3 { // 2 steps + desfecho
		t.Fatalf("esperava 3 eventos, veio %d: %+v", len(eventos), eventos)
	}
	if eventos[0].Step != 1 || eventos[1].Step != 2 || eventos[2].Step != 0 {
		t.Fatalf("sequência de eventos inesperada: %+v", eventos)
	}
	if eventos[1].Causation != "step-1" || eventos[2].Attrs["status"] != string(wf.RunSucceeded) {
		t.Fatalf("correlação/atributos inesperados: %+v", eventos)
	}
	for _, ev := range eventos {
		if ev.Correlation != string(run.ID) || ev.Schema != wf.EventSchema {
			t.Fatalf("envelope fora do contrato: %+v", ev)
		}
	}
}
