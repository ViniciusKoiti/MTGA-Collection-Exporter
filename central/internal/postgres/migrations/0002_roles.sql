-- Central roles (OpenSpec add-central-go-platform, task 2.2). All roles
-- are NOLOGIN groups: production creates LOGIN users as members. Runtime
-- roles never own tables, are never superusers and never BYPASSRLS.

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_owner') THEN
        CREATE ROLE central_owner NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_migrator') THEN
        CREATE ROLE central_migrator NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_api') THEN
        CREATE ROLE central_api NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_telemetry') THEN
        CREATE ROLE central_telemetry NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_worker') THEN
        CREATE ROLE central_worker NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_operations') THEN
        CREATE ROLE central_operations NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'central_backup') THEN
        CREATE ROLE central_backup NOLOGIN;
    END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO central_api, central_telemetry,
    central_worker, central_operations, central_backup;
GRANT CREATE, USAGE ON SCHEMA public TO central_owner, central_migrator;

-- API: serve manifests, enroll installations, record consent and audit.
-- It never reads raw accepted telemetry events and never does DDL.
GRANT SELECT ON sources, snapshots, artifacts TO central_api;
GRANT SELECT, INSERT, UPDATE ON installations, consent_receipts,
    deletion_requests TO central_api;
GRANT INSERT ON audit TO central_api;

-- Telemetry: append-only ingestion scoped by installation.
GRANT SELECT ON installations, consent_receipts TO central_telemetry;
GRANT INSERT ON telemetry_batches, accepted_events TO central_telemetry;

-- Worker: publication and aggregation jobs; may expire raw events.
GRANT SELECT, INSERT, UPDATE ON jobs, outbox, artifacts, snapshots,
    sources, aggregates TO central_worker;
GRANT SELECT, INSERT, DELETE ON accepted_events TO central_worker;
GRANT SELECT ON telemetry_batches TO central_worker;
GRANT INSERT ON audit TO central_worker;

-- Operations: read-only evidence, never raw telemetry events.
GRANT SELECT ON sources, snapshots, artifacts, jobs, outbox, audit,
    aggregates, deletion_requests TO central_operations;

-- Backup: full read for dumps, nothing else.
GRANT pg_read_all_data TO central_backup;

-- Sequences used by the writing roles.
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public
    TO central_api, central_telemetry, central_worker;
