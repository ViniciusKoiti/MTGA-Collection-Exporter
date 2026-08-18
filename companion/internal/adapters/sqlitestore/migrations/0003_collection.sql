CREATE TABLE snapshots (
    id              TEXT PRIMARY KEY,
    schema          TEXT    NOT NULL,
    source          TEXT    NOT NULL,
    source_instance TEXT    NOT NULL DEFAULT '',
    observed_at     TEXT    NOT NULL,
    imported_at     TEXT    NOT NULL,
    is_latest       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_snapshots_latest ON snapshots (is_latest);

CREATE TABLE snapshot_entries (
    snapshot_id TEXT    NOT NULL REFERENCES snapshots (id),
    idx         INTEGER NOT NULL,
    printing    TEXT    NOT NULL DEFAULT '',
    arena       INTEGER NOT NULL DEFAULT 0,
    oracle      TEXT    NOT NULL DEFAULT '',
    name        TEXT    NOT NULL DEFAULT '',
    set_code    TEXT    NOT NULL DEFAULT '',
    quantity    INTEGER NOT NULL,
    unresolved  INTEGER NOT NULL DEFAULT 0,
    raw         TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (snapshot_id, idx)
);

CREATE TABLE snapshot_diagnostics (
    snapshot_id TEXT    NOT NULL REFERENCES snapshots (id),
    idx         INTEGER NOT NULL,
    code        TEXT    NOT NULL,
    detail      TEXT    NOT NULL DEFAULT '',
    severity    TEXT    NOT NULL,
    PRIMARY KEY (snapshot_id, idx)
);
