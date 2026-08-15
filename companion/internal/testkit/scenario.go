package testkit

import (
	"time"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// Decisao é uma decisão de aprovação prevista pelo cenário, indexada pelo
// nome do efeito do preview exato.
type Decisao struct {
	Conceder bool
	Validade time.Duration
}

// Scenario descreve uma execução determinística: grafo de produção,
// fixtures de estado, política, decisões previstas e decoradores de falha
// (tarefa 5.1 parcial; a forma JSON versionada chega depois).
type Scenario struct {
	Nome      string
	Graph     wf.Identity
	Registrar func(*wf.Registry, *memory.Clock) error
	Estado    wf.State
	Policy    wf.Policy
	// Aprovacoes indexa decisões previstas por EffectPreview.Effect.
	Aprovacoes map[string]Decisao
	Inicio     time.Time
	// DecorarStore envolve o run store de produção com decoradores de
	// falha (ex.: CrashAposCheckpoints); nil mantém o store puro.
	DecorarStore func(wf.RunStore) wf.RunStore
	// CaminhoBanco, quando definido, troca o store in-memory pelo SQLite
	// de produção (migrações aplicadas na composição — perfil da 5.2).
	CaminhoBanco string
	// Atividade, quando definida, executa via activity.Launcher (entry
	// point de produção) em vez de chamar o engine diretamente.
	Atividade string
	// Inventario é obrigatório quando Atividade é usada.
	Inventario *activity.Registry
}
