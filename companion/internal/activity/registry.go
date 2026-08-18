// Package activity mantém o inventário executável de atividades (tarefas
// 1.3/2.7 do OpenSpec add-graph-workflow-harness): todo comando externo ou
// job agendado resolve para um grafo registrado ou para uma isenção de
// consulta pura documentada. Entry points que não constam aqui não podem
// ser expostos — o teste de arquitetura correspondente chega com o
// roteamento Wails (tarefa 4.6).
package activity

import (
	"errors"
	"fmt"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Kind separa atividade orquestrada de consulta pura isenta.
type Kind string

// Tipos de atividade válidos.
const (
	KindGraph     Kind = "graph"
	KindPureQuery Kind = "pure_query"
)

// Entry é uma atividade registrada.
type Entry struct {
	Name          string
	Kind          Kind
	Graph         wf.Identity // preenchido apenas em KindGraph
	Justification string      // obrigatório em KindPureQuery
}

// ErrActivityInvalida cobre registros e consultas rejeitados.
var ErrActivityInvalida = errors.New("activity: registro inválido")

// Registry é montado na composição do processo, nunca em runtime.
type Registry struct {
	entries map[string]Entry
}

// New cria o inventário vazio.
func New() *Registry {
	return &Registry{entries: make(map[string]Entry)}
}

// RegisterGraph liga uma atividade externa a um grafo versionado.
func (r *Registry) RegisterGraph(name string, graph wf.Identity) error {
	if name == "" || graph.Kind == "" || graph.Version < 1 {
		return fmt.Errorf("%w: %q -> %s", ErrActivityInvalida, name, graph)
	}
	return r.insere(Entry{Name: name, Kind: KindGraph, Graph: graph})
}

// RegisterPureQuery isenta uma consulta pura; a justificativa é obrigatória
// e vira documentação auditável do inventário.
func (r *Registry) RegisterPureQuery(name, justification string) error {
	if name == "" || justification == "" {
		return fmt.Errorf("%w: consulta pura %q exige justificativa", ErrActivityInvalida, name)
	}
	return r.insere(Entry{Name: name, Kind: KindPureQuery, Justification: justification})
}

func (r *Registry) insere(e Entry) error {
	if _, existe := r.entries[e.Name]; existe {
		return fmt.Errorf("%w: atividade %q duplicada", ErrActivityInvalida, e.Name)
	}
	r.entries[e.Name] = e
	return nil
}

// Resolve devolve a atividade registrada; desconhecida é erro — não existe
// caminho implícito que contorne um grafo.
func (r *Registry) Resolve(name string) (Entry, error) {
	e, ok := r.entries[name]
	if !ok {
		return Entry{}, fmt.Errorf("%w: atividade %q não registrada", ErrActivityInvalida, name)
	}
	return e, nil
}

// Validate confirma que toda atividade de grafo aponta para uma definição
// realmente registrada no runtime; roda na composição e nos testes.
func (r *Registry) Validate(grafos *wf.Registry) error {
	for _, e := range r.entries {
		if e.Kind != KindGraph {
			continue
		}
		if _, err := grafos.Get(e.Graph); err != nil {
			return fmt.Errorf("%w: atividade %q aponta para grafo ausente %s",
				ErrActivityInvalida, e.Name, e.Graph)
		}
	}
	return nil
}
