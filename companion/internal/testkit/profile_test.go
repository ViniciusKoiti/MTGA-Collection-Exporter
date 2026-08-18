package testkit

import (
	"path/filepath"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/activity"
	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// TestPerfilSQLiteComEntryPointDeAtividade executa o cenário pelo entry
// point de produção (activity.Launcher) sobre o SQLite migrado — o perfil
// de composição completo da tarefa 5.2 — e usa as asserções da 5.5.
func TestPerfilSQLiteComEntryPointDeAtividade(t *testing.T) {
	inventario := activity.New()
	if err := inventario.RegisterGraph("export-approved",
		wf.Identity{Kind: "approved-export", Version: 1}); err != nil {
		t.Fatalf("inventário: %v", err)
	}
	sc := cenarioExport(true)
	sc.CaminhoBanco = filepath.Join(t.TempDir(), "harness.db")
	sc.Atividade = "export-approved"
	sc.Inventario = inventario

	h, err := New(sc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	defer func() { _ = h.Close() }()
	res, err := h.Run(t.Context())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	for nome, verifica := range map[string]error{
		"desfecho":   res.AssertDesfecho(wf.RunSucceeded, "done"),
		"transições": res.AssertTransicoes("valida", "exporta", "exporta"),
		"tentativas": res.AssertTentativasDoNo("exporta", 2),
		"checkpoint": res.AssertCheckpointConsistente(),
		"efeitos":    res.AssertEfeitosPendentes(0),
		"orçamento":  res.AssertOrcamento(8),
		"eventos":    res.AssertEventos(5), // 3 steps + waiting + desfecho
	} {
		if verifica != nil {
			t.Errorf("%s: %v", nome, verifica)
		}
	}
}

// TestAtividadeForaDoInventarioNaoExecuta prova que o harness respeita o
// mesmo bloqueio de bypass do produto.
func TestAtividadeForaDoInventarioNaoExecuta(t *testing.T) {
	sc := cenarioExport(true)
	sc.Atividade = "atividade-fantasma"
	sc.Inventario = activity.New()
	h, err := New(sc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	if _, err := h.inicia(t.Context()); err == nil {
		t.Fatal("atividade fora do inventário não pode executar")
	}
}
