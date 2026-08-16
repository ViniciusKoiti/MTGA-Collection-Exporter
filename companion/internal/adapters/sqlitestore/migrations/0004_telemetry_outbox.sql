-- Durable local telemetry outbox (central OpenSpec task 5.2): events
-- stay on the device until the central acknowledges them. Unacked
-- rows survive restarts. attrs_json carries allowlisted attrs only.
-- NOTE: no semicolons inside comments -- the migration runner splits
-- statements on them.
CREATE TABLE IF NOT EXISTS telemetry_outbox (
    seq        INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    at         TEXT NOT NULL,
    attrs_json TEXT NOT NULL DEFAULT '{}',
    acked      INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_telemetry_outbox_pending
    ON telemetry_outbox (acked, seq)
