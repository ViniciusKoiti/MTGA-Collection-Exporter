package decks

import (
	"testing"
	"time"
)

func standard(t *testing.T) (Standard, time.Time) {
	t.Helper()
	agora := time.Unix(1_700_000_000, 0)
	legais := map[string]bool{"lightning strike": true, "abrade": true}
	return StandardPadrao(legais, agora.Add(-time.Hour)), agora
}

func deckStandard(t *testing.T, strikes int, side []Entry) Deck {
	t.Helper()
	deck, err := NewDeck("Mono Vermelho", []Entry{
		{Name: "Lightning Strike", Quantity: strikes},
		{Name: "Mountain", Quantity: 60 - strikes},
	}, side)
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	return deck
}

func TestDeckLegalPassaCitandoRuleset(t *testing.T) {
	ruleset, agora := standard(t)
	v := ruleset.Validate(deckStandard(t, 4, []Entry{{Name: "Abrade", Quantity: 2}}), agora)
	if !v.Legal || v.CatalogoStale || v.Ruleset.String() != "standard/v1" {
		t.Fatalf("veredicto inesperado: %+v", v)
	}
}

func TestViolacoesTemCodigosEstaveis(t *testing.T) {
	ruleset, agora := standard(t)
	deck, err := NewDeck("Quebrado", []Entry{
		{Name: "Lightning Strike", Quantity: 5}, // cópias excedidas
		{Name: "Carta Proibida", Quantity: 4},   // ilegal
		{Name: "Mountain", Quantity: 30},        // básico isento de cópias
	}, []Entry{{Name: "Mountain", Quantity: 16}}) // sideboard grande; básico isento de cópias
	if err != nil {
		t.Fatalf("deck: %v", err)
	}
	v := ruleset.Validate(deck, agora) // 39 cartas: deck pequeno também
	if v.Legal {
		t.Fatal("deck com violações não pode ser legal")
	}
	codigos := map[string]int{}
	for _, p := range v.Problemas {
		codigos[p.Code]++
	}
	esperados := map[string]int{
		"deck_pequeno": 1, "sideboard_grande": 1,
		"copias_excedidas": 1, "carta_ilegal": 1,
	}
	for code, quantos := range esperados {
		if codigos[code] != quantos {
			t.Fatalf("esperava %d %s, veio %+v", quantos, code, v.Problemas)
		}
	}
	if codigos["copias_excedidas"] == 1 {
		for _, p := range v.Problemas {
			if p.Code == "copias_excedidas" && p.Carta != "lightning strike" {
				t.Fatalf("cópias excedidas na carta errada: %+v", p)
			}
		}
	}
}

func TestCatalogoVelhoESinalizadoSemInventarLegalidade(t *testing.T) {
	ruleset, agora := standard(t)
	ruleset.CatalogoAtualizado = agora.Add(-30 * 24 * time.Hour)
	v := ruleset.Validate(deckStandard(t, 4, nil), agora)
	if !v.CatalogoStale {
		t.Fatal("catálogo com 30 dias deveria ser sinalizado como stale")
	}
	if !v.Legal { // legalidade continua sendo a do catálogo conhecido
		t.Fatalf("stale não muda o veredicto de legalidade: %+v", v)
	}
}
