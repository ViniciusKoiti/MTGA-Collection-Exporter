// Package approvedexport implementa o grafo approved-export v1 (tarefa 4.3
// do OpenSpec add-graph-workflow-harness): preview imutável do efeito,
// aprovação expirável via effect gate, escrita idempotente pelo outbox e
// evidência de conclusão no journal de eventos.
package approvedexport

import (
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Deps são os ports consumidos pelos nós.
type Deps struct {
	Snapshots ports.SnapshotStore
	Exporter  ports.Exporter
	Outbox    wf.Outbox
	Clock     ports.Clock
}

// Estado é o estado tipado do run. O preview do efeito é imutável: o hash
// do payload exato é fixado em `carrega` e a aprovação vale só para ele —
// payload diferente exige nova aprovação.
type Estado struct {
	Formato  ports.ExportFormat
	Destino  string
	Snapshot collection.SnapshotID
	Hash     string // sha256 do payload exato projetado
	Bytes    int
	EfeitoID string // chave de idempotência enfileirada no outbox
}

// estadoDe normaliza o estado recebido pelo nó, aplicando defaults.
func estadoDe(st any) Estado {
	estado, _ := st.(Estado)
	if estado.Formato == "" {
		estado.Formato = ports.ExportJSON
	}
	if estado.Destino == "" {
		estado.Destino = "mtga_collection." + string(estado.Formato)
	}
	return estado
}
