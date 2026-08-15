package decks

import (
	"fmt"
	"time"
)

// Problema é uma violação estável encontrada pela validação de ruleset.
type Problema struct {
	Code   string // estável: deck_pequeno, copias_excedidas, ...
	Carta  string // vazio quando o problema é do deck inteiro
	Detail string
}

// Veredicto é o resultado da validação, sempre citando o ruleset exato e
// se o catálogo de legalidade estava velho no momento da checagem.
type Veredicto struct {
	Ruleset       RulesetID
	Problemas     []Problema
	CatalogoStale bool
	Legal         bool
}

// Standard é o ruleset Standard versionado (tarefa 5.3 do companion).
// A legalidade vem injetada do catálogo com o instante de atualização;
// o domínio nunca busca dados sozinho.
type Standard struct {
	ID                 RulesetID
	MinDeck            int
	MaxCopias          int
	MaxSideboard       int
	TerrenosBasicos    map[string]bool // isentos do limite de cópias
	Legais             map[string]bool // nomes normalizados legais no formato
	CatalogoAtualizado time.Time
	ValidadeCatalogo   time.Duration
}

// StandardPadrao devolve o ruleset com os limites oficiais do formato.
func StandardPadrao(legais map[string]bool, atualizado time.Time) Standard {
	basicos := map[string]bool{"plains": true, "island": true, "swamp": true,
		"mountain": true, "forest": true, "wastes": true}
	return Standard{
		ID: RulesetID{Format: "standard", Version: 1}, MinDeck: 60, MaxCopias: 4,
		MaxSideboard: 15, TerrenosBasicos: basicos, Legais: legais,
		CatalogoAtualizado: atualizado, ValidadeCatalogo: 7 * 24 * time.Hour,
	}
}

// Validate aplica tamanho de deck, limite de cópias, sideboard e
// legalidade; devolve problemas com códigos estáveis e sinaliza catálogo
// velho sem inventar legalidade.
func (r Standard) Validate(deck Deck, agora time.Time) Veredicto {
	v := Veredicto{Ruleset: r.ID}
	if agora.Sub(r.CatalogoAtualizado) > r.ValidadeCatalogo {
		v.CatalogoStale = true
	}
	if total := deck.TotalPrincipal(); total < r.MinDeck {
		v.Problemas = append(v.Problemas, Problema{Code: "deck_pequeno",
			Detail: fmt.Sprintf("%d cartas; mínimo %d", total, r.MinDeck)})
	}
	lado := 0
	for _, e := range deck.Sideboard {
		lado += e.Quantity
	}
	if lado > r.MaxSideboard {
		v.Problemas = append(v.Problemas, Problema{Code: "sideboard_grande",
			Detail: fmt.Sprintf("%d cartas; máximo %d", lado, r.MaxSideboard)})
	}
	copias := map[string]int{}
	for _, grupo := range [][]Entry{deck.Main, deck.Sideboard} {
		for _, e := range grupo {
			copias[chaveNome(e.Name)] += e.Quantity
		}
	}
	for chave, total := range copias {
		if r.TerrenosBasicos[chave] {
			continue
		}
		if total > r.MaxCopias {
			v.Problemas = append(v.Problemas, Problema{Code: "copias_excedidas",
				Carta: chave, Detail: fmt.Sprintf("%d cópias; máximo %d", total, r.MaxCopias)})
		}
		if !r.Legais[chave] {
			v.Problemas = append(v.Problemas, Problema{Code: "carta_ilegal", Carta: chave,
				Detail: "fora da lista de legalidade do catálogo"})
		}
	}
	v.Legal = len(v.Problemas) == 0
	return v
}
