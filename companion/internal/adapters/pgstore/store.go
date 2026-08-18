// Package pgstore is the development/staging PostgreSQL run-store
// adapter (OpenSpec add-graph-workflow-harness, task 3.6). It implements
// the exact same optimistic contract as the in-memory and SQLite
// adapters and passes the shared storetest suite against a real
// PostgreSQL via Testcontainers — skipped locally without Docker,
// mandatory in CI.
package pgstore

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx"
)

const schema = `
CREATE TABLE IF NOT EXISTS wf_runs (
    id            TEXT PRIMARY KEY,
    graph_kind    TEXT    NOT NULL,
    graph_version INTEGER NOT NULL,
    status        TEXT    NOT NULL,
    current_node  TEXT    NOT NULL,
    state_json    TEXT    NOT NULL,
    outcome       TEXT    NOT NULL DEFAULT '',
    run_version   INTEGER NOT NULL,
    started_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL
);
CREATE TABLE IF NOT EXISTS wf_steps (
    seq         BIGSERIAL PRIMARY KEY,
    run_id      TEXT    NOT NULL,
    idx         INTEGER NOT NULL,
    node        TEXT    NOT NULL,
    attempt     INTEGER NOT NULL,
    outcome     TEXT    NOT NULL DEFAULT '',
    err         TEXT    NOT NULL DEFAULT '',
    started_at  TEXT    NOT NULL,
    finished_at TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_wf_steps_run ON wf_steps (run_id, seq);`

// Store implements wf.RunStore over PostgreSQL.
type Store struct {
	db *sql.DB
}

// Open refuses production identities, connects, applies the forward
// schema and returns the store.
func Open(dsn string) (*Store, error) {
	if err := GuardDSN(dsn); err != nil {
		return nil, err
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("pgstore: open: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pgstore: migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the connection pool.
func (s *Store) Close() error { return s.db.Close() }

func instant(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func parseInstant(text string) (time.Time, error) {
	if text == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, text)
}
