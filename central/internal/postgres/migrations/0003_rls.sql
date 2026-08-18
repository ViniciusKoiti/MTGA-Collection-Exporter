-- Row Level Security (OpenSpec add-central-go-platform, task 2.6).
-- The ingestion path (central_telemetry) is scoped to ONE installation
-- per transaction via SET LOCAL app.installation_id; with the setting
-- unset every row is invisible. Worker keeps cross-installation access
-- for aggregation and expiry; API keeps lookup access for enrollment
-- and token authentication; backup reads for dumps.

-- RLS scopes rows, but the base privilege must exist: telemetry reads
-- its own batches/events (0002 granted INSERT only).
GRANT SELECT ON telemetry_batches, accepted_events TO central_telemetry;

ALTER TABLE installations     ENABLE ROW LEVEL SECURITY;
ALTER TABLE consent_receipts  ENABLE ROW LEVEL SECURITY;
ALTER TABLE telemetry_batches ENABLE ROW LEVEL SECURITY;
ALTER TABLE accepted_events   ENABLE ROW LEVEL SECURITY;

-- Telemetry: strictly its own installation, read and write.
CREATE POLICY telemetry_own_installation ON installations
    FOR SELECT TO central_telemetry
    USING (id = current_setting('app.installation_id', true));

CREATE POLICY telemetry_own_consent ON consent_receipts
    FOR SELECT TO central_telemetry
    USING (installation_id = current_setting('app.installation_id', true));

CREATE POLICY telemetry_own_batches ON telemetry_batches
    TO central_telemetry
    USING (installation_id = current_setting('app.installation_id', true))
    WITH CHECK (installation_id = current_setting('app.installation_id', true));

CREATE POLICY telemetry_own_events ON accepted_events
    TO central_telemetry
    USING (batch_id IN (SELECT id FROM telemetry_batches
        WHERE installation_id = current_setting('app.installation_id', true)))
    WITH CHECK (batch_id IN (SELECT id FROM telemetry_batches
        WHERE installation_id = current_setting('app.installation_id', true)));

-- API: enrollment and token lookup need the full tables.
CREATE POLICY api_full_installations ON installations
    TO central_api USING (true) WITH CHECK (true);
CREATE POLICY api_full_consent ON consent_receipts
    TO central_api USING (true) WITH CHECK (true);

-- Worker: aggregation and expiry are cross-installation by design.
CREATE POLICY worker_full_batches ON telemetry_batches
    TO central_worker USING (true) WITH CHECK (true);
CREATE POLICY worker_full_events ON accepted_events
    TO central_worker USING (true) WITH CHECK (true);

-- Backup: read-only dumps of everything.
CREATE POLICY backup_read_installations ON installations
    FOR SELECT TO central_backup USING (true);
CREATE POLICY backup_read_consent ON consent_receipts
    FOR SELECT TO central_backup USING (true);
CREATE POLICY backup_read_batches ON telemetry_batches
    FOR SELECT TO central_backup USING (true);
CREATE POLICY backup_read_events ON accepted_events
    FOR SELECT TO central_backup USING (true);
