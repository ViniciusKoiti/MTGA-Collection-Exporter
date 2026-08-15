// Package arch contém os testes de arquitetura do módulo companion
// (tarefas 1.4 e 8.1 do OpenSpec add-graph-workflow-harness).
package arch

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestArquivosAte100Linhas aplica o gate de 100 linhas físicas a todo
// arquivo Go mantido manualmente (exclusões documentadas: nenhuma ainda).
func TestArquivosAte100Linhas(t *testing.T) {
	for arquivo, linhas := range arquivosGo(t) {
		if n := len(linhas); n > 100 {
			t.Errorf("%s tem %d linhas; o limite do módulo é 100", arquivo, n)
		}
	}
}

// TestComandosNaoContornamOInventarioDeAtividades garante que nenhum
// binário (cmd/*) fala com o runtime de grafos diretamente: comandos e
// jobs externos executam apenas via internal/activity.Launcher, que
// resolve o inventário antes de iniciar qualquer grafo (tarefa 2.7).
func TestComandosNaoContornamOInventarioDeAtividades(t *testing.T) {
	fset := token.NewFileSet()
	for arquivo := range arquivosGo(t) {
		normalizado := strings.ReplaceAll(arquivo, "\\", "/")
		if !strings.Contains(normalizado, "/cmd/") {
			continue
		}
		tree, err := parser.ParseFile(fset, arquivo, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse de %s: %v", arquivo, err)
		}
		for _, imp := range tree.Imports {
			caminho := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(caminho, "/internal/workflow") {
				t.Errorf("%s importa o runtime diretamente; use internal/activity.Launcher", arquivo)
			}
		}
	}
}

// TestNucleoIndependeDeAdaptadores garante que workflow e domínio não
// importam Wails, SQLite, PostgreSQL, MCP nem SDKs de modelo: esses
// pacotes dependem só da stdlib e de outros pacotes do núcleo — nunca de
// adapters, testkit ou frontend (tarefas 1.4 do harness e 2.5 do companion).
func TestNucleoIndependeDeAdaptadores(t *testing.T) {
	fset := token.NewFileSet()
	for arquivo := range arquivosGo(t) {
		normalizado := strings.ReplaceAll(arquivo, "\\", "/")
		if !strings.Contains(normalizado, "internal/workflow/") &&
			!strings.Contains(normalizado, "internal/domain/") {
			continue
		}
		tree, err := parser.ParseFile(fset, arquivo, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse de %s: %v", arquivo, err)
		}
		for _, imp := range tree.Imports {
			caminho := strings.Trim(imp.Path.Value, `"`)
			proprio := strings.HasPrefix(caminho,
				"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/")
			if strings.Contains(caminho, ".") && !proprio { // fora da stdlib e do módulo
				t.Errorf("%s importa dependência externa proibida: %s", arquivo, caminho)
			}
			for _, vetado := range []string{"/internal/adapters/", "/internal/testkit", "/frontend"} {
				if strings.Contains(caminho, vetado) {
					t.Errorf("%s: núcleo não pode depender de %s", arquivo, caminho)
				}
			}
		}
	}
}
