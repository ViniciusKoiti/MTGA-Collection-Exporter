// Package devscenario implementa o grafo development-scenario v1 (tarefa
// 4.5 do OpenSpec add-graph-workflow-harness): valida um cenário JSON
// estrito, executa UM grafo alvo pelo harness de produção, compara a
// evidência com a asserção declarada e produz um relatório redigido.
// É o veículo de execução do futuro dev-mcp.
package devscenario

import (
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/testkit"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// Registrar registra o grafo alvo e suas dependências fake no registry do
// cenário; o catálogo é montado em código na composição de desenvolvimento.
type Registrar func(*wf.Registry, *memory.Clock) error

// Deps é o catálogo de cenários executáveis: grafos alvo conhecidos e as
// políticas opcionais de cada um. Cenários JSON só referenciam entradas
// deste catálogo — nunca definem transições.
type Deps struct {
	Registrars map[string]Registrar // chave: kind do grafo alvo
	Policies   map[string]wf.Policy // política opcional por kind
}

// Estado é o estado tipado do run; carrega apenas dados serializáveis.
type Estado struct {
	CenarioJSON string
	Parsed      testkit.ScenarioJSON
	Desfecho    string // status/outcome do run alvo
	Evidencia   string // evidência normalizada do run alvo
	Relatorio   string // relatório final redigido
}

// estadoDe normaliza o estado recebido pelo nó.
func estadoDe(st any) Estado {
	estado, _ := st.(Estado)
	return estado
}
