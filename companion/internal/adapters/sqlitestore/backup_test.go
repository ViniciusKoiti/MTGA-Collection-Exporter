package sqlitestore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// TestCheckpointCorrompidoFalhaLimpo prova que estado adulterado no banco
// vira erro explícito de leitura — nunca um estado silenciosamente vazio.
func TestCheckpointCorrompidoFalhaLimpo(t *testing.T) {
	store := abre(t)
	interrompido(t, store, "run-corrompido")
	if _, err := store.db.Exec(
		`UPDATE runs SET state_json = '{corrompido' WHERE id = 'run-corrompido'`); err != nil {
		t.Fatalf("adulterar: %v", err)
	}
	if _, err := store.Get(t.Context(), "run-corrompido"); err == nil {
		t.Fatal("checkpoint corrompido deveria falhar a leitura explicitamente")
	}
	engine, _ := engineSobre(t, store, store, "w1")
	if _, err := engine.Recover(t.Context(), "run-corrompido", "w1", time.Minute); err == nil {
		t.Fatal("recuperação sobre checkpoint corrompido deveria falhar limpo")
	}
}

// TestBackupERollbackRestauramConsistencia cobre backup e rollback da
// tarefa 3.7: cópia fria do arquivo, perda do banco corrente e restauração
// devolvem exatamente o estado do momento do backup.
func TestBackupERollbackRestauramConsistencia(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "runs.db")
	backup := caminho + ".bak"

	primeiro, err := Open(caminho)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	interrompido(t, primeiro, "run-do-backup")
	if err := primeiro.Close(); err != nil { // fecha para checkpoint frio do WAL
		t.Fatalf("close: %v", err)
	}
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatalf("ler banco: %v", err)
	}
	if err := os.WriteFile(backup, conteudo, 0o644); err != nil {
		t.Fatalf("backup: %v", err)
	}

	segundo, err := Open(caminho)
	if err != nil {
		t.Fatalf("reabrir: %v", err)
	}
	interrompido(t, segundo, "run-pos-backup") // escrita que o rollback descarta
	if err := segundo.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := os.Remove(caminho); err != nil { // banco corrente perdido
		t.Fatalf("remover: %v", err)
	}
	restaurado, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("ler backup: %v", err)
	}
	if err := os.WriteFile(caminho, restaurado, 0o644); err != nil {
		t.Fatalf("restaurar: %v", err)
	}

	terceiro, err := Open(caminho) // migrações idempotentes sobre o restaurado
	if err != nil {
		t.Fatalf("abrir restaurado: %v", err)
	}
	defer func() { _ = terceiro.Close() }()
	if _, err := terceiro.Get(t.Context(), "run-do-backup"); err != nil {
		t.Fatalf("estado do backup deveria voltar: %v", err)
	}
	if _, err := terceiro.Get(t.Context(), "run-pos-backup"); err == nil {
		t.Fatal("rollback deveria descartar escritas posteriores ao backup")
	}
	run, _ := terceiro.Get(t.Context(), "run-do-backup")
	if run.Status != wf.RunActive || run.Current != "valida" {
		t.Fatalf("checkpoint restaurado divergente: %+v", run)
	}
}
