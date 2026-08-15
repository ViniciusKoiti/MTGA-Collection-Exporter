package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow/memory"
)

// ScenarioSchema é a versão corrente da forma JSON de cenários (5.1).
const ScenarioSchema = "scenario/v1"

// Limites estritos da forma JSON: cenário e fixture têm teto de tamanho.
const (
	maxScenarioBytes = 64 << 10
	maxFixtureBytes  = 32 << 10
)

// ScenarioJSON é a forma serializada de um cenário: dados apenas — o grafo
// e a política continuam em Go, ligados em Materializar.
type ScenarioJSON struct {
	Schema string `json:"schema"`
	Nome   string `json:"nome"`
	Graph  struct {
		Kind    string `json:"kind"`
		Version int    `json:"version"`
	} `json:"graph"`
	Estado     json.RawMessage `json:"estado,omitempty"`
	Aprovacoes map[string]struct {
		Conceder         bool  `json:"conceder"`
		ValidadeSegundos int64 `json:"validade_segundos"`
	} `json:"aprovacoes,omitempty"`
	InicioUnix int64 `json:"inicio_unix"`
	// Esperado é a asserção declarada do cenário (status/outcome finais);
	// o grafo development-scenario compara e reprova divergência.
	Esperado *struct {
		Status  string `json:"status"`
		Outcome string `json:"outcome"`
	} `json:"esperado,omitempty"`
}

// ParseScenario valida a forma JSON estritamente: campo desconhecido,
// tamanho excedido, schema errado ou identidade inválida são erros.
func ParseScenario(dados []byte) (ScenarioJSON, error) {
	if len(dados) > maxScenarioBytes {
		return ScenarioJSON{}, fmt.Errorf("testkit: cenário excede %d bytes", maxScenarioBytes)
	}
	decodificador := json.NewDecoder(bytes.NewReader(dados))
	decodificador.DisallowUnknownFields()
	var sc ScenarioJSON
	if err := decodificador.Decode(&sc); err != nil {
		return ScenarioJSON{}, fmt.Errorf("testkit: cenário inválido: %w", err)
	}
	if decodificador.More() {
		return ScenarioJSON{}, fmt.Errorf("testkit: conteúdo extra após o cenário")
	}
	if sc.Schema != ScenarioSchema {
		return ScenarioJSON{}, fmt.Errorf("testkit: schema %q não suportado", sc.Schema)
	}
	if sc.Nome == "" || sc.Graph.Kind == "" || sc.Graph.Version < 1 || sc.InicioUnix < 1 {
		return ScenarioJSON{}, fmt.Errorf("testkit: cenário %q com identidade incompleta", sc.Nome)
	}
	if len(sc.Estado) > maxFixtureBytes {
		return ScenarioJSON{}, fmt.Errorf("testkit: fixture excede %d bytes", maxFixtureBytes)
	}
	return sc, nil
}

// Materializar liga a forma JSON ao código de produção: registrador de
// grafos e política vêm de Go; dados e decisões vêm do JSON.
func (sc ScenarioJSON) Materializar(
	registrar func(*wf.Registry, *memory.Clock) error,
	policy wf.Policy,
) (Scenario, error) {
	var estado wf.State
	if len(sc.Estado) > 0 {
		if err := json.Unmarshal(sc.Estado, &estado); err != nil {
			return Scenario{}, fmt.Errorf("testkit: fixture malformada: %w", err)
		}
	}
	aprovacoes := make(map[string]Decisao, len(sc.Aprovacoes))
	for efeito, decisao := range sc.Aprovacoes {
		aprovacoes[efeito] = Decisao{
			Conceder: decisao.Conceder,
			Validade: time.Duration(decisao.ValidadeSegundos) * time.Second,
		}
	}
	return Scenario{
		Nome:       sc.Nome,
		Graph:      wf.Identity{Kind: wf.Kind(sc.Graph.Kind), Version: sc.Graph.Version},
		Registrar:  registrar,
		Estado:     estado,
		Policy:     policy,
		Aprovacoes: aprovacoes,
		Inicio:     time.Unix(sc.InicioUnix, 0).UTC(),
	}, nil
}
