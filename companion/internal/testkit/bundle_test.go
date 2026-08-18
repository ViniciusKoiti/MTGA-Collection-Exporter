package testkit

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var atualizaGolden = flag.Bool("update", false, "regrava os arquivos golden")

// TestBundleDiagnosticoBateComGoldenDePrivacidade congela o formato do
// bundle: qualquer campo novo aparece como diff do golden e força revisão
// de privacidade explícita (tarefa 6.7).
func TestBundleDiagnosticoBateComGoldenDePrivacidade(t *testing.T) {
	res := executa(t, cenarioExport(true))
	bundle := res.DiagnosticBundle()
	caminho := filepath.Join("testdata", "bundle_approved_export.golden")
	if *atualizaGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("testdata: %v", err)
		}
		if err := os.WriteFile(caminho, []byte(bundle), 0o644); err != nil {
			t.Fatalf("gravar golden: %v", err)
		}
	}
	esperado, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatalf("golden ausente (rode com -update): %v", err)
	}
	if bundle != string(esperado) {
		t.Fatalf("bundle divergiu do golden de privacidade:\n--- got:\n%s--- want:\n%s",
			bundle, esperado)
	}
}

// TestBundleNaoContemCamposProibidos garante a redação do pacote inteiro.
func TestBundleNaoContemCamposProibidos(t *testing.T) {
	res := executa(t, cenarioExport(true))
	bundle := res.DiagnosticBundle()
	proibidos := []string{
		"decks/a.txt", // alvo do efeito
		"cartas",      // fixture de estado
		"h1",          // hash de payload cru
		"C:\\", "prompt", "password",
	}
	for _, fragmento := range proibidos {
		if strings.Contains(bundle, fragmento) {
			t.Fatalf("bundle contém fragmento proibido %q:\n%s", fragmento, bundle)
		}
	}
}
