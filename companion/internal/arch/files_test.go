package arch

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// raizDoModulo sobe até encontrar o go.mod do módulo central.
func raizDoModulo(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		pai := filepath.Dir(dir)
		if pai == dir {
			t.Fatal("go.mod do módulo central não encontrado")
		}
		dir = pai
	}
}

// arquivosGo devolve o conteúdo em linhas de cada arquivo Go do módulo.
func arquivosGo(t *testing.T) map[string][]string {
	t.Helper()
	arquivos := make(map[string][]string)
	raiz := raizDoModulo(t)
	err := filepath.WalkDir(raiz, func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(caminho, ".go") {
			return nil
		}
		conteudo, err := os.ReadFile(caminho)
		if err != nil {
			return err
		}
		texto := strings.TrimRight(string(conteudo), "\n")
		arquivos[caminho] = strings.Split(texto, "\n")
		return nil
	})
	if err != nil {
		t.Fatalf("varredura do módulo: %v", err)
	}
	if len(arquivos) == 0 {
		t.Fatal("nenhum arquivo Go encontrado")
	}
	return arquivos
}
