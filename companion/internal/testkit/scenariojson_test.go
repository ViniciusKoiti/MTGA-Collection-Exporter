package testkit

import (
	"strings"
	"testing"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

func cenarioJSONValido() string {
	return `{
		"schema": "scenario/v1",
		"nome": "approved-export-json",
		"graph": {"kind": "approved-export", "version": 1},
		"estado": {"cartas": 3},
		"aprovacoes": {"export-write": {"conceder": true, "validade_segundos": 3600}},
		"inicio_unix": 1700000000
	}`
}

func TestParseScenarioAceitaFormaValidaEExecuta(t *testing.T) {
	parsed, err := ParseScenario([]byte(cenarioJSONValido()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	base := cenarioExport(true) // reusa grafo e política de produção
	sc, err := parsed.Materializar(base.Registrar, base.Policy)
	if err != nil {
		t.Fatalf("materializar: %v", err)
	}
	res := executa(t, sc)
	if err := res.AssertDesfecho(wf.RunSucceeded, "done"); err != nil {
		t.Fatal(err)
	}
}

func TestParseScenarioRejeitaFormasInvalidas(t *testing.T) {
	casos := map[string]string{
		"campo desconhecido": strings.Replace(cenarioJSONValido(),
			`"nome"`, `"intruso": 1, "nome"`, 1),
		"schema errado": strings.Replace(cenarioJSONValido(),
			"scenario/v1", "scenario/v9", 1),
		"versão zero": strings.Replace(cenarioJSONValido(),
			`"version": 1`, `"version": 0`, 1),
		"sem início": strings.Replace(cenarioJSONValido(),
			`"inicio_unix": 1700000000`, `"inicio_unix": 0`, 1),
		"conteúdo extra": cenarioJSONValido() + `{"outro": 1}`,
		"fixture gigante": strings.Replace(cenarioJSONValido(),
			`{"cartas": 3}`, `{"x": "`+strings.Repeat("a", 40<<10)+`"}`, 1),
	}
	for nome, dados := range casos {
		if _, err := ParseScenario([]byte(dados)); err == nil {
			t.Errorf("%s: parse deveria falhar", nome)
		}
	}
}
