## ADDED Requirements

### Requirement: Database roles are least privilege
The platform SHALL use separate owner, migrator, API, telemetry, worker, operations,
and backup roles, and runtime roles MUST NOT own tables or use `BYPASSRLS`.

#### Scenario: Telemetry role reads catalog governance
- **WHEN** the telemetry role attempts a query outside its grants
- **THEN** PostgreSQL denies it and the denial is observable without exposing data

### Requirement: Migrations are controlled
Forward SQL migrations SHALL run through a dedicated command and migrator role with
preflight, lock, version, backup policy, and post-migration verification.

#### Scenario: API starts with unsupported schema
- **WHEN** the database version is outside the binary compatibility range
- **THEN** readiness fails and the API does not serve application traffic

### Requirement: Tenant-scoped data is isolated
Traceable installation and consent tables SHALL enforce principal scoping in queries
and database policy, with default-deny behavior for missing context.

#### Scenario: Installation reads another installation
- **WHEN** an installation-scoped operation supplies another installation identity
- **THEN** it returns no data and cannot infer whether the other identity exists

### Requirement: Secrets are externally managed
Database, object-store, signing, OIDC, and backup secrets SHALL come from an approved
secret store, remain absent from code and artifacts, and support rotation.

#### Scenario: Repository is scanned
- **WHEN** CI inspects source, history, images, and artifacts
- **THEN** no production credential or private signing key is present

### Requirement: Retention and deletion are table specific
Every table SHALL have documented purpose, classification, retention, deletion,
backup interaction, and legal-hold behavior before production data is stored.

#### Scenario: Deletion reaches a backup
- **WHEN** traceable data exists in an immutable backup
- **THEN** restore procedures reapply tombstones before restored services become available

### Requirement: Backup and restore are proven
The platform SHALL create encrypted verified backups and SHALL test restore against
documented RPO and RTO before production launch and on a recurring schedule.

#### Scenario: Restore drill runs
- **WHEN** operators restore the latest eligible backup into isolation
- **THEN** migrations, checks, tombstones, and service smoke tests pass within RTO
