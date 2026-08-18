package workflow

import (
	"context"
	"errors"
	"testing"
	"time"
)

// noFake é um nó trivial para testes estruturais.
type noFake struct{ id NodeID }

func (n noFake) ID() NodeID { return n.id }
func (n noFake) Execute(context.Context, State) (State, OutcomeCode, error) {
	return nil, "ok", nil
}

// grafoValido monta: inicio -ok-> fim -ok-> terminal "done".
func grafoValido() Definition {
	return Definition{
		Identity: Identity{Kind: "collection-sync", Version: 1},
		Initial:  "inicio",
		Nodes:    map[NodeID]Node{"inicio": noFake{"inicio"}, "fim": noFake{"fim"}},
		Transitions: map[TransitionKey]Target{
			{From: "inicio", Outcome: "ok"}: {Next: "fim"},
			{From: "fim", Outcome: "ok"}:    {Terminal: "done"},
		},
		Recovery: RecoveryResume,
	}
}

func TestRegistraGrafoValidoComDefaults(t *testing.T) {
	r := NewRegistry(Limits{})
	if err := r.Register(grafoValido()); err != nil {
		t.Fatalf("registro deveria passar: %v", err)
	}
	d, err := r.Get(Identity{Kind: "collection-sync", Version: 1})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if d.Limits != DefaultLimits() {
		t.Fatalf("limites zerados deveriam virar defaults: %+v", d.Limits)
	}
}

func TestRejeitaDefinicoesInvalidas(t *testing.T) {
	casos := map[string]func(*Definition){
		"kind vazio":          func(d *Definition) { d.Identity.Kind = "" },
		"versão zero":         func(d *Definition) { d.Identity.Version = 0 },
		"inicial inexistente": func(d *Definition) { d.Initial = "fantasma" },
		"recovery inválida":   func(d *Definition) { d.Recovery = "talvez" },
		"transição para nó inexistente": func(d *Definition) {
			d.Transitions[TransitionKey{From: "inicio", Outcome: "ok"}] = Target{Next: "fantasma"}
		},
		"alvo ambíguo": func(d *Definition) {
			d.Transitions[TransitionKey{From: "fim", Outcome: "ok"}] = Target{Next: "inicio", Terminal: "done"}
		},
		"nó sem saída": func(d *Definition) {
			delete(d.Transitions, TransitionKey{From: "fim", Outcome: "ok"})
		},
		"nó inalcançável": func(d *Definition) {
			d.Nodes["ilha"] = noFake{"ilha"}
			d.Transitions[TransitionKey{From: "ilha", Outcome: "ok"}] = Target{Terminal: "done"}
		},
		"sem terminal alcançável": func(d *Definition) {
			d.Transitions[TransitionKey{From: "fim", Outcome: "ok"}] = Target{Next: "inicio"}
		},
		"limites acima do teto": func(d *Definition) {
			d.Limits = Limits{MaxSteps: 99, MaxToolCalls: 1, MaxRepeatedCalls: 1, ActiveDeadline: time.Second}
		},
	}
	for nome, mutacao := range casos {
		d := grafoValido()
		mutacao(&d)
		if err := NewRegistry(Limits{}).Register(d); !errors.Is(err, ErrInvalidDefinition) {
			t.Errorf("%s: esperava ErrInvalidDefinition, veio %v", nome, err)
		}
	}
}

func TestRejeitaIdentidadeDuplicadaEConsultaDesconhecida(t *testing.T) {
	r := NewRegistry(Limits{})
	if err := r.Register(grafoValido()); err != nil {
		t.Fatalf("primeiro registro: %v", err)
	}
	if err := r.Register(grafoValido()); !errors.Is(err, ErrDuplicateIdentity) {
		t.Fatalf("duplicata deveria falhar, veio: %v", err)
	}
	if _, err := r.Get(Identity{Kind: "x", Version: 9}); !errors.Is(err, ErrUnknownIdentity) {
		t.Fatalf("identidade desconhecida deveria falhar, veio: %v", err)
	}
}
