package activity

import (
	"context"
	"fmt"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Launcher é a ÚNICA porta pela qual comandos externos e jobs iniciam
// trabalho: resolve a atividade no inventário e executa o grafo pinado.
// O teste de arquitetura impede cmd/* de importar internal/workflow, então
// não existe caminho de execução que contorne o inventário (tarefa 2.7).
type Launcher struct {
	inventario *Registry
	engine     *wf.Engine
}

// NewLauncher valida o inventário contra o registry de grafos na
// composição: atividade apontando para grafo ausente reprova o boot.
func NewLauncher(inventario *Registry, grafos *wf.Registry, engine *wf.Engine) (*Launcher, error) {
	if err := inventario.Validate(grafos); err != nil {
		return nil, err
	}
	return &Launcher{inventario: inventario, engine: engine}, nil
}

// Run executa a atividade nomeada; consulta pura não passa por aqui.
func (l *Launcher) Run(ctx context.Context, nome string, estado wf.State) (wf.Run, error) {
	entrada, err := l.inventario.Resolve(nome)
	if err != nil {
		return wf.Run{}, err
	}
	if entrada.Kind != KindGraph {
		return wf.Run{}, fmt.Errorf("%w: %q é consulta pura e não inicia grafo",
			ErrActivityInvalida, nome)
	}
	return l.engine.Start(ctx, entrada.Graph, estado)
}
