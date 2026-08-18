// Package collectionsync implementa o grafo collection-sync v1 (tarefa 4.1
// do OpenSpec add-graph-workflow-harness): detecção de fonte, observação,
// normalização preservando não resolvidas, validação, commit de snapshot
// imutável e projeções de compatibilidade — tudo por ports do companion.
package collectionsync

import (
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// Deps são os ports que os nós consomem; a composição injeta produção ou
// os fakes do harness — o grafo é idêntico nos dois casos.
type Deps struct {
	Source    ports.CollectionSource
	Catalog   ports.Catalog
	Snapshots ports.SnapshotStore
	Exporter  ports.Exporter
	Clock     ports.Clock
}

// Estado é o estado tipado do run, serializável para checkpoint. Ele
// carrega apenas o necessário entre nós; fixtures do perfil determinístico
// cabem no orçamento de payload do grafo.
type Estado struct {
	Observacao    collection.Observation
	Entradas      []collection.Entry
	Diagnosticos  []collection.Diagnostic
	Snapshot      collection.SnapshotID
	Resolvidas    int
	NaoResolvidas int
	Projecoes     map[string]int // formato -> bytes projetados
}

// estadoDe normaliza o estado recebido pelo nó (nil no primeiro nó).
func estadoDe(st any) Estado {
	if estado, ok := st.(Estado); ok {
		return estado
	}
	return Estado{}
}
