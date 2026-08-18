// Package sqlitestore implementa os ports de persistência do runtime de
// grafos sobre SQLite local (driver Go puro modernc.org/sqlite), com
// migrações forward embutidas e checkpoints transacionais (tarefas
// 3.1/3.2 do OpenSpec add-graph-workflow-harness).
package sqlitestore

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migracoes embed.FS

// migrar aplica as migrações forward pendentes, cada uma em uma transação,
// registrando a versão aplicada em schema_migrations.
func migrar(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("sqlitestore: bootstrap de migrações: %w", err)
	}
	entradas, err := migracoes.ReadDir("migrations")
	if err != nil {
		return err
	}
	nomes := make([]string, 0, len(entradas))
	for _, e := range entradas {
		nomes = append(nomes, e.Name())
	}
	sort.Strings(nomes) // ordem forward determinística
	for _, nome := range nomes {
		if err := aplicar(db, nome); err != nil {
			return err
		}
	}
	return nil
}

func aplicar(db *sql.DB, nome string) error {
	var existe int
	err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, nome).
		Scan(&existe)
	if err != nil {
		return err
	}
	if existe > 0 {
		return nil
	}
	conteudo, err := migracoes.ReadFile("migrations/" + nome)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, comando := range strings.Split(string(conteudo), ";") {
		if strings.TrimSpace(comando) == "" {
			continue
		}
		if _, err := tx.Exec(comando); err != nil {
			return fmt.Errorf("sqlitestore: migração %s: %w", nome, err)
		}
	}
	agora := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, applied_at)
		VALUES (?, ?)`, nome, agora); err != nil {
		return err
	}
	return tx.Commit()
}
