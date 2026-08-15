CREATE TABLE audit (
    seq         INTEGER PRIMARY KEY AUTOINCREMENT,
    correlation TEXT NOT NULL,
    tool        TEXT NOT NULL,
    args_hash   TEXT NOT NULL,
    outcome     TEXT NOT NULL,
    at          TEXT NOT NULL
);

CREATE INDEX idx_audit_correlation ON audit (correlation, seq);
