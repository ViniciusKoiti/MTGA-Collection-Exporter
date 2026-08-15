// Package arch contém os testes de arquitetura do módulo central
// (tarefas 1.2 e 6.5 do OpenSpec add-central-go-platform).
package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestArquivosAte100Linhas aplica o gate de tamanho a todo arquivo Go.
func TestArquivosAte100Linhas(t *testing.T) {
	for arquivo, linhas := range arquivosGo(t) {
		if n := len(linhas); n > 100 {
			t.Errorf("%s tem %d linhas; o limite do módulo é 100", arquivo, n)
		}
	}
}

// TestGoroutinesECanaisSomenteNoPacoteDono garante que apenas
// internal/platform/concurrency cria goroutines e canais em código de
// produção: goroutine solta e canal sem dono são proibidos por spec.
func TestGoroutinesECanaisSomenteNoPacoteDono(t *testing.T) {
	fset := token.NewFileSet()
	for arquivo := range arquivosGo(t) {
		if strings.HasSuffix(arquivo, "_test.go") || ehPacoteDono(arquivo) {
			continue
		}
		tree, err := parser.ParseFile(fset, arquivo, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse de %s: %v", arquivo, err)
		}
		ast.Inspect(tree, func(no ast.Node) bool {
			switch n := no.(type) {
			case *ast.GoStmt:
				t.Errorf("%s: goroutine fora do pacote de concorrência (linha %d)",
					arquivo, fset.Position(n.Pos()).Line)
			case *ast.ChanType:
				t.Errorf("%s: canal declarado fora do pacote de concorrência (linha %d)",
					arquivo, fset.Position(n.Pos()).Line)
			}
			return true
		})
	}
}

// TestConcorrenciaNaoUsaNumerosMagicos garante que o pacote de concorrência
// nunca escolhe limites por conta própria: workers e buffers chegam sempre
// por parâmetro, derivados de config.Budgets.
func TestConcorrenciaNaoUsaNumerosMagicos(t *testing.T) {
	fset := token.NewFileSet()
	for arquivo := range arquivosGo(t) {
		if !ehPacoteDono(arquivo) || strings.HasSuffix(arquivo, "_test.go") {
			continue
		}
		tree, err := parser.ParseFile(fset, arquivo, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse de %s: %v", arquivo, err)
		}
		ast.Inspect(tree, func(no ast.Node) bool {
			chamada, ok := no.(*ast.CallExpr)
			if !ok || len(chamada.Args) < 2 {
				return true
			}
			if ident, ok := chamada.Fun.(*ast.Ident); ok && ident.Name == "make" {
				if _, ehCanal := chamada.Args[0].(*ast.ChanType); ehCanal {
					if lit, fixo := chamada.Args[1].(*ast.BasicLit); fixo && lit.Value != "0" {
						t.Errorf("%s: capacidade de canal fixa em código (linha %d)",
							arquivo, fset.Position(lit.Pos()).Line)
					}
				}
			}
			return true
		})
	}
}

func ehPacoteDono(arquivo string) bool {
	normalizado := strings.ReplaceAll(arquivo, "\\", "/")
	return strings.Contains(normalizado, "internal/platform/concurrency/")
}
