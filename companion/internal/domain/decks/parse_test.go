package decks

import (
	"strings"
	"testing"
)

const fixtureArena = `About
Name Mono Vermelho

Deck
4 Lightning Strike (DMU) 137
2 Ilha
20 Mountain (UNF) 240

Sideboard
2 Abrade (VOW) 139
1 Carta Desconhecida Do Futuro
`

func TestParseArenaTextCompleto(t *testing.T) {
	deck, err := ParseArenaText(fixtureArena, "fallback")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if deck.Name != "Mono Vermelho" || len(deck.Main) != 3 || len(deck.Sideboard) != 2 {
		t.Fatalf("estrutura inesperada: %+v", deck)
	}
	if deck.Main[0].Set != "DMU" || deck.Main[0].Numero != "137" || deck.Main[0].Quantity != 4 {
		t.Fatalf("linha com set/número mal parseada: %+v", deck.Main[0])
	}
	if deck.Main[1].Name != "Ilha" || deck.Main[1].Set != "" { // nome localizado sem set
		t.Fatalf("nome localizado mal parseado: %+v", deck.Main[1])
	}
	if deck.Sideboard[1].Name != "Carta Desconhecida Do Futuro" {
		t.Fatalf("carta desconhecida deveria ser preservada: %+v", deck.Sideboard[1])
	}
}

func TestParseRejeitaLinhaMalformadaComNumero(t *testing.T) {
	casos := []string{
		"Deck\nLightning Strike (DMU) 137\n", // sem quantidade
		"Deck\n0 Mountain\n",                 // quantidade zero
		"Deck\n4\n",                          // sem nome
	}
	for _, texto := range casos {
		if _, err := ParseArenaText(texto, "x"); err == nil ||
			!strings.Contains(err.Error(), "linha") {
			t.Errorf("deveria falhar citando a linha: %q -> %v", texto, err)
		}
	}
	if _, err := ParseArenaText("", "x"); err == nil {
		t.Error("texto vazio não tem cartas principais e deveria falhar")
	}
}

func TestRoundTripFormatacao(t *testing.T) {
	deck, err := ParseArenaText(fixtureArena, "fallback")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	reparseado, err := ParseArenaText(deck.ArenaText(), "fallback")
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if reparseado.Name != deck.Name || len(reparseado.Main) != len(deck.Main) ||
		len(reparseado.Sideboard) != len(deck.Sideboard) {
		t.Fatalf("round-trip divergente:\n%+v\n%+v", deck, reparseado)
	}
	for i := range deck.Main {
		if reparseado.Main[i] != deck.Main[i] {
			t.Fatalf("entrada %d divergiu no round-trip: %+v vs %+v",
				i, deck.Main[i], reparseado.Main[i])
		}
	}
}
