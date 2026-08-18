package decks

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// linhaCarta reconhece "4 Nome da Carta (SET) 123" e "4 Nome da Carta";
// nomes localizados (acentos, hífens) são preservados como vieram.
var linhaCarta = regexp.MustCompile(`^(\d+)\s+(.+?)(?:\s+\(([A-Z0-9]{2,6})\)\s+(\S+))?$`)

// ParseArenaText converte o texto de deck exportado pelo MTG Arena em um
// Deck validado (tarefa 5.1 do companion). Seções reconhecidas: About/
// Name, Deck e Sideboard. Linha fora do formato é erro com o número da
// linha — nunca descarte silencioso.
func ParseArenaText(texto, nomeDefault string) (Deck, error) {
	nome := nomeDefault
	var principal, sideboard []Entry
	secao := "deck"
	esperaNome := false
	for numero, linha := range strings.Split(strings.ReplaceAll(texto, "\r\n", "\n"), "\n") {
		limpa := strings.TrimSpace(linha)
		switch {
		case limpa == "":
			continue
		case esperaNome && strings.HasPrefix(limpa, "Name "):
			nome = strings.TrimPrefix(limpa, "Name ")
			esperaNome = false
			continue
		case limpa == "About":
			esperaNome = true
			continue
		case limpa == "Deck":
			secao = "deck"
			continue
		case limpa == "Sideboard":
			secao = "sideboard"
			continue
		}
		entrada, err := parseLinha(limpa)
		if err != nil {
			return Deck{}, fmt.Errorf("decks: linha %d malformada: %w", numero+1, err)
		}
		if secao == "sideboard" {
			sideboard = append(sideboard, entrada)
		} else {
			principal = append(principal, entrada)
		}
	}
	return NewDeck(nome, principal, sideboard)
}

func parseLinha(linha string) (Entry, error) {
	grupos := linhaCarta.FindStringSubmatch(linha)
	if grupos == nil {
		return Entry{}, fmt.Errorf("%q não segue o formato Arena", linha)
	}
	quantidade, err := strconv.Atoi(grupos[1])
	if err != nil || quantidade < 1 {
		return Entry{}, fmt.Errorf("quantidade inválida em %q", linha)
	}
	return Entry{
		Name:     grupos[2],
		Quantity: quantidade,
		Set:      grupos[3],
		Numero:   grupos[4],
	}, nil
}

// ArenaText formata o deck de volta ao texto Arena, preservando set e
// número quando conhecidos (round-trip estável com o parser).
func (d Deck) ArenaText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "About\nName %s\n\nDeck\n", d.Name)
	escreve := func(entradas []Entry) {
		for _, e := range entradas {
			if e.Set != "" && e.Numero != "" {
				fmt.Fprintf(&b, "%d %s (%s) %s\n", e.Quantity, e.Name, e.Set, e.Numero)
			} else {
				fmt.Fprintf(&b, "%d %s\n", e.Quantity, e.Name)
			}
		}
	}
	escreve(d.Main)
	if len(d.Sideboard) > 0 {
		b.WriteString("\nSideboard\n")
		escreve(d.Sideboard)
	}
	return b.String()
}
