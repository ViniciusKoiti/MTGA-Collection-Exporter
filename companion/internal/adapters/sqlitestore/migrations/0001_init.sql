CREATE TABLE runs (
    id            TEXT PRIMARY KEY,
    graph_kind    TEXT    NOT NULL,
    graph_version INTEGER NOT NULL,
    status        TEXT    NOT NULL,
    current_node  TEXT    NOT NULL,
    state_json    TEXT    NOT NULL,
    outcome       TEXT    NOT NULL DEFAULT '',
    run_version   INTEGER NOT NULL,
    lease_owner   TEXT    NOT NULL DEFAULT '',
    lease_until   TEXT    NOT NULL DEFAULT '',
    started_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL
);

CREATE TABLE steps (
    seq         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id      TEXT    NOT NULL REFERENCES runs (id),
    idx         INTEGER NOT NULL,
    node        TEXT    NOT NULL,
    attempt     INTEGER NOT NULL,
    outcome     TEXT    NOT NULL DEFAULT '',
    err         TEXT    NOT NULL DEFAULT '',
    started_at  TEXT    NOT NULL,
    finished_at TEXT    NOT NULL
);

CREATE INDEX idx_steps_run ON steps (run_id, seq);

CREATE TABLE approvals (
    run_id       TEXT    NOT NULL,
    hash         TEXT    NOT NULL,
    effect       TEXT    NOT NULL,
    target       TEXT    NOT NULL,
    payload_hash TEXT    NOT NULL,
    decided      INTEGER NOT NULL DEFAULT 0,
    granted      INTEGER NOT NULL DEFAULT 0,
    expires_at   TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (run_id, hash)
);

CREATE TABLE outbox (
    id           TEXT PRIMARY KEY,
    run_id       TEXT    NOT NULL,
    effect       TEXT    NOT NULL,
    target       TEXT    NOT NULL,
    payload_hash TEXT    NOT NULL,
    enqueued_at  TEXT    NOT NULL,
    acked        INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE events (
    seq           INTEGER PRIMARY KEY AUTOINCREMENT,
    schema        TEXT    NOT NULL,
    run_id        TEXT    NOT NULL,
    step          INTEGER NOT NULL,
    graph_kind    TEXT    NOT NULL,
    graph_version INTEGER NOT NULL,
    correlation   TEXT    NOT NULL,
    causation     TEXT    NOT NULL DEFAULT '',
    at            TEXT    NOT NULL,
    outcome       TEXT    NOT NULL,
    duration_ms   INTEGER NOT NULL,
    attrs_json    TEXT    NOT NULL DEFAULT '{}'
);

CREATE TABLE graph_identities (
    kind          TEXT    NOT NULL,
    version       INTEGER NOT NULL,
    registered_at TEXT    NOT NULL,
    PRIMARY KEY (kind, version)
);
