// Package metarecommend implementa o grafo meta-deck-recommendation v1
// (tarefa 4.2 do OpenSpec add-graph-workflow-harness): carrega snapshot e
// catálogo de meta, compara posse, ranqueia por montabilidade e devolve
// lacunas com evidência de proveniência.
package metarecommend

import (
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Deps são os ports consumidos pelos nós.
type Deps struct {
	Snapshots ports.SnapshotStore
	Meta      ports.MetaDeckCatalog
	// MaxLacunas limita os nomes de cartas faltantes citados por deck.
	MaxLacunas int
}

// Recomendacao é uma linha do ranking: determinística e com evidência.
type Recomendacao struct {
	Nome          string
	Fonte         string // proveniência do meta deck
	Montabilidade int    // percentual 0-100 de cartas possuídas
	Faltantes     int
	Lacunas       []string // principais cartas faltantes, limitadas
}

// Estado é o estado tipado do run.
type Estado struct {
	Formato       string
	Snapshot      collection.SnapshotID
	Avaliados     int
	Recomendacoes []Recomendacao
}

// estadoDe normaliza o estado recebido pelo nó, aplicando defaults.
func estadoDe(st any) Estado {
	estado, _ := st.(Estado)
	if estado.Formato == "" {
		estado.Formato = "standard"
	}
	return estado
}
