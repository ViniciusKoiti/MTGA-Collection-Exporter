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

// TestWorkflowIndependeDeAdaptadores garante que as definições de workflow
// não importam Wails, SQLite, PostgreSQL, MCP nem SDKs de modelo: o pacote
// internal/workflow só pode depender da biblioteca padrão.
func TestWorkflowIndependeDeAdaptadores(t *testing.T) {
	fset := token.NewFileSet()
	for arquivo := range arquivosGo(t) {
		normalizado := strings.ReplaceAll(arquivo, "\\", "/")
		if !strings.Contains(normalizado, "internal/workflow/") {
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
		}
	}
}
