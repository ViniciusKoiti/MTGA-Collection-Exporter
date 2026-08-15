-- Central schema v1 (OpenSpec add-central-go-platform, task 2.1):
-- catalogs, installations, consent, telemetry, aggregates, deletion,
-- jobs, outbox and audit. Forward-only; roles and RLS arrive in 0002+.

CREATE TABLE sources (
    id          TEXT PRIMARY KEY,
    kind        TEXT        NOT NULL,
    approved    BOOLEAN     NOT NULL DEFAULT FALSE,
    rights_note TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE snapshots (
    id            TEXT PRIMARY KEY,
    source_id     TEXT        NOT NULL REFERENCES sources (id),
    schema_name   TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE artifacts (
    id           TEXT PRIMARY KEY,
    snapshot_id  TEXT        NOT NULL REFERENCES snapshots (id),
    object_key   TEXT        NOT NULL,
    sha256       TEXT        NOT NULL,
    size_bytes   BIGINT      NOT NULL,
    signed_by    TEXT        NOT NULL DEFAULT '',
    is_current   BOOLEAN     NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ
);

CREATE TABLE installations (
    id              TEXT PRIMARY KEY,
    token_hash      TEXT        NOT NULL,
    deletion_hash   TEXT        NOT NULL,
    client_version  TEXT        NOT NULL DEFAULT '',
    revoked         BOOLEAN     NOT NULL DEFAULT FALSE,
    enrolled_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE consent_receipts (
    id              BIGSERIAL PRIMARY KEY,
    installation_id TEXT        NOT NULL REFERENCES installations (id),
    purpose         TEXT        NOT NULL,
    version         INTEGER     NOT NULL,
    granted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at      TIMESTAMPTZ
);

CREATE TABLE telemetry_batches (
    id              TEXT PRIMARY KEY,
    installation_id TEXT        NOT NULL REFERENCES installations (id),
    sequence        BIGINT      NOT NULL,
    accepted_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (installation_id, sequence)
);

CREATE TABLE accepted_events (
    id         BIGSERIAL PRIMARY KEY,
    batch_id   TEXT        NOT NULL REFERENCES telemetry_batches (id),
    name       TEXT        NOT NULL,
    attrs      JSONB       NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE aggregates (
    metric     TEXT        NOT NULL,
    bucket     TIMESTAMPTZ NOT NULL,
    value      BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (metric, bucket)
);

CREATE TABLE deletion_requests (
    id              BIGSERIAL PRIMARY KEY,
    installation_id TEXT        NOT NULL,
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    proof           TEXT        NOT NULL DEFAULT ''
);

CREATE TABLE jobs (
    id              TEXT PRIMARY KEY,
    kind            TEXT        NOT NULL,
    idempotency_key TEXT        NOT NULL UNIQUE,
    status          TEXT        NOT NULL DEFAULT 'pending',
    attempts        INTEGER     NOT NULL DEFAULT 0,
    lease_owner     TEXT        NOT NULL DEFAULT '',
    lease_until     TIMESTAMPTZ,
    payload         JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_jobs_claim ON jobs (status, lease_until);

CREATE TABLE outbox (
    id           TEXT PRIMARY KEY,
    kind         TEXT        NOT NULL,
    payload      JSONB       NOT NULL DEFAULT '{}',
    enqueued_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    acked        BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE TABLE audit (
    id          BIGSERIAL PRIMARY KEY,
    actor       TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    subject     TEXT        NOT NULL DEFAULT '',
    at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
