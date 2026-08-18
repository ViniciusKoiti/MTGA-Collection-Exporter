package decks

import (
	"sort"
	"strings"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
)

// OwnershipLine é o resultado de posse de uma carta do deck: exigido,
// possuído e faltante. WildcardRelevant é o faltante coberto por
// wildcards; hoje igual a Missing — o refinamento por raridade chega com
// o catálogo Scryfall (tarefa 3.6 do companion).
type OwnershipLine struct {
	Name             string
	Required         int
	Owned            int
	Missing          int
	WildcardRelevant int
}

// Ownership é a comparação completa do deck contra um snapshot.
type Ownership struct {
	Snapshot     collection.SnapshotID
	Lines        []OwnershipLine
	TotalMissing int
	Complete     bool
}

// CompareOwnership compara o deck (principal + sideboard) com o snapshot
// selecionado (tarefa 5.2). A posse é funcional: impressões diferentes da
// mesma carta somam, casadas por nome normalizado; cartas não resolvidas
// do snapshot não contam como posse.
func CompareOwnership(deck Deck, snap collection.Snapshot) Ownership {
	possuidas := make(map[string]int)
	for _, entrada := range snap.Entries {
		if !entrada.Unresolved {
			possuidas[chaveNome(entrada.Identity.Name)] += entrada.Quantity
		}
	}
	exigidas := make(map[string]int)
	nomes := make(map[string]string) // chave normalizada -> nome original
	for _, grupo := range [][]Entry{deck.Main, deck.Sideboard} {
		for _, e := range grupo {
			chave := chaveNome(e.Name)
			exigidas[chave] += e.Quantity
			nomes[chave] = e.Name
		}
	}
	resultado := Ownership{Snapshot: snap.ID}
	for chave, exigido := range exigidas {
		possuido := min(possuidas[chave], exigido)
		faltante := exigido - possuido
		resultado.Lines = append(resultado.Lines, OwnershipLine{
			Name: nomes[chave], Required: exigido, Owned: possuido,
			Missing: faltante, WildcardRelevant: faltante,
		})
		resultado.TotalMissing += faltante
	}
	sort.Slice(resultado.Lines, func(i, j int) bool {
		return resultado.Lines[i].Name < resultado.Lines[j].Name
	})
	resultado.Complete = resultado.TotalMissing == 0
	return resultado
}

// chaveNome normaliza o nome para o casamento funcional.
func chaveNome(nome string) string {
	return strings.ToLower(strings.TrimSpace(nome))
}
