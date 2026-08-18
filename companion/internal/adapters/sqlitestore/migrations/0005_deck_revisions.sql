-- Durable deck revisions (companion task 5.6 follow-up): every saved
-- revision keeps its deck as canonical JSON plus the snapshot and
-- ruleset identity that made its verdict reproducible.
CREATE TABLE deck_revisions (
    id          TEXT PRIMARY KEY,
    deck_name   TEXT NOT NULL,
    deck_json   TEXT NOT NULL,
    snapshot_id TEXT NOT NULL,
    ruleset     TEXT NOT NULL,
    saved_at    TEXT NOT NULL
);
CREATE INDEX idx_deck_revisions_name ON deck_revisions (deck_name, saved_at, id)
