package decks

import "testing"

func TestNewDeckValidaInvariantes(t *testing.T) {
	principal := []Entry{{Name: "Carta A", Arena: 1, Quantity: 4}}
	deck, err := NewDeck("Mono Azul", principal, []Entry{{Name: "Carta B", Quantity: 2}})
	if err != nil || deck.TotalPrincipal() != 4 {
		t.Fatalf("deck válido deveria construir: %v / %+v", err, deck)
	}
	principal[0].Quantity = 99
	if deck.Main[0].Quantity != 4 {
		t.Fatal("deck deveria copiar defensivamente as entradas")
	}
	casos := map[string]func() (Deck, error){
		"sem nome":       func() (Deck, error) { return NewDeck("", principal, nil) },
		"sem principais": func() (Deck, error) { return NewDeck("X", nil, nil) },
		"quantidade zero": func() (Deck, error) {
			return NewDeck("X", []Entry{{Name: "A"}}, nil)
		},
		"entrada sem nome": func() (Deck, error) {
			return NewDeck("X", []Entry{{Quantity: 1}}, nil)
		},
	}
	for nome, constroi := range casos {
		if _, err := constroi(); err == nil {
			t.Errorf("%s: deveria falhar", nome)
		}
	}
}

func TestRulesetIDValida(t *testing.T) {
	if err := (RulesetID{Format: "standard", Version: 1}).Valida(); err != nil {
		t.Fatalf("ruleset válido: %v", err)
	}
	if err := (RulesetID{}).Valida(); err == nil {
		t.Fatal("ruleset vazio deveria falhar")
	}
	if got := (RulesetID{Format: "standard", Version: 3}).String(); got != "standard/v3" {
		t.Fatalf("identidade textual inesperada: %s", got)
	}
}
