// Package decks define o domínio de decks e rulesets do companion
// (tarefa 2.2 do OpenSpec introduce-agentic-go-companion). Só stdlib.
package decks

import (
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// Entry é uma linha de deck: nome como veio do texto Arena, Arena ID
// quando resolvido e quantidade.
type Entry struct {
	Name     string
	Arena    collection.ArenaID
	Quantity int
}

// Deck é um deck no formato Arena: principal e sideboard.
type Deck struct {
	Name      string
	Main      []Entry
	Sideboard []Entry
}

// NewDeck valida os invariantes básicos do formato: nome presente,
// quantidades positivas e pelo menos uma carta principal.
func NewDeck(name string, main, sideboard []Entry) (Deck, error) {
	if name == "" {
		return Deck{}, fmt.Errorf("decks: deck sem nome")
	}
	if len(main) == 0 {
		return Deck{}, fmt.Errorf("decks: deck %q sem cartas principais", name)
	}
	for _, grupo := range [][]Entry{main, sideboard} {
		for _, e := range grupo {
			if e.Quantity < 1 {
				return Deck{}, fmt.Errorf("decks: %q com quantidade inválida %d", e.Name, e.Quantity)
			}
			if e.Name == "" {
				return Deck{}, fmt.Errorf("decks: entrada sem nome no deck %q", name)
			}
		}
	}
	return Deck{
		Name:      name,
		Main:      append([]Entry(nil), main...),
		Sideboard: append([]Entry(nil), sideboard...),
	}, nil
}

// TotalPrincipal soma as cartas do deck principal.
func (d Deck) TotalPrincipal() int {
	total := 0
	for _, e := range d.Main {
		total += e.Quantity
	}
	return total
}

// RulesetID identifica a versão exata das regras usadas em uma validação;
// resultados de legalidade sempre citam o ruleset que os produziu.
type RulesetID struct {
	Format  string // ex.: "standard"
	Version int
}

func (r RulesetID) String() string {
	return fmt.Sprintf("%s/v%d", r.Format, r.Version)
}

// Valida confere os invariantes da identidade de ruleset.
func (r RulesetID) Valida() error {
	if r.Format == "" || r.Version < 1 {
		return fmt.Errorf("decks: ruleset inválido %q v%d", r.Format, r.Version)
	}
	return nil
}
