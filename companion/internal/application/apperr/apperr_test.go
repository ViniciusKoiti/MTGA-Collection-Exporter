package apperr

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestApresentavelNaoVazaCausaTecnica(t *testing.T) {
	causa := fmt.Errorf("open C:\\Users\\usuario\\AppData\\mtga_collection.json: acesso negado")
	err := New(CodeSourceUnavailable, "collection.sync", causa)
	seguro := err.Apresentavel()
	for _, proibido := range []string{"C:\\", "AppData", "mtga_collection.json"} {
		if strings.Contains(seguro, proibido) {
			t.Fatalf("apresentação vazou detalhe técnico %q: %s", proibido, seguro)
		}
	}
	if !strings.Contains(seguro, string(CodeSourceUnavailable)) ||
		!strings.Contains(seguro, "collection.sync") {
		t.Fatalf("apresentação deveria trazer código e operação: %s", seguro)
	}
	if !strings.Contains(err.Error(), "acesso negado") {
		t.Fatal("log local deveria preservar a causa para diagnóstico")
	}
}

func TestCodeOfExtraiCodigoEstavel(t *testing.T) {
	base := New(CodeSnapshotNotFound, "collection.get", nil)
	embrulhado := fmt.Errorf("camada acima: %w", base)
	if CodeOf(embrulhado) != CodeSnapshotNotFound {
		t.Fatalf("código deveria sobreviver ao embrulho: %s", CodeOf(embrulhado))
	}
	if CodeOf(errors.New("qualquer coisa")) != CodeInternal {
		t.Fatal("erro não tipado é interno por definição")
	}
	if !errors.Is(embrulhado, base) {
		t.Fatal("errors.Is deveria enxergar o erro tipado")
	}
}
