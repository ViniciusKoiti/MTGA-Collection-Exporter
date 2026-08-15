package sqlitestore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // driver SQLite em Go puro

	wf "github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/workflow"
)

// Store implementa wf.RunStore sobre SQLite com a mesma semântica de
// versão otimista do adapter in-memory (contrato storetest).
type Store struct {
	db *sql.DB
}

// Open abre (ou cria) o banco no caminho dado e aplica as migrações.
func Open(caminho string) (*Store, error) {
	db, err := sql.Open("sqlite", caminho)
	if err != nil {
		return nil, fmt.Errorf("sqlitestore: abrir %s: %w", caminho, err)
	}
	// Escritor único local: evita SQLITE_BUSY em testes concorrentes.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrar(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close fecha a conexão com o banco.
func (s *Store) Close() error { return s.db.Close() }

// serializaEstado fixa o estado como JSON; estado não serializável já foi
// barrado pelo engine (validaPayload), então aqui é erro de programação.
func serializaEstado(st wf.State) (string, error) {
	b, err := json.Marshal(st)
	if err != nil {
		return "", fmt.Errorf("sqlitestore: estado não serializável: %w", err)
	}
	return string(b), nil
}

func desserializaEstado(texto string) (wf.State, error) {
	var st wf.State
	if err := json.Unmarshal([]byte(texto), &st); err != nil {
		return nil, fmt.Errorf("sqlitestore: estado corrompido: %w", err)
	}
	return st, nil
}

func formataInstante(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseInstante(texto string) (time.Time, error) {
	if texto == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, texto)
}
