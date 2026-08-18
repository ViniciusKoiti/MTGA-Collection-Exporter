## Why

The local companion needs a real central service for versioned meta catalogs and
consented aggregate telemetry, but user collections and raw logs must remain local.
A separately deployable Go backend establishes this trust boundary and provides a
measurable path to scale without introducing distributed infrastructure prematurely.

## What Changes

- Add a central Go module with stateless HTTPS API and background worker binaries.
- Publish immutable, hash-addressed catalog snapshots through object storage/CDN and
  expose small manifest endpoints with freshness and compatibility metadata.
- Ingest explicit opt-in telemetry as bounded, idempotent batches after schema,
  privacy, rate, and installation-policy validation.
- Store catalog metadata, provider provenance, consent receipts, aggregate metrics,
  deletion state, jobs, and audit evidence in PostgreSQL.
- Use bounded Go concurrency, request-scoped cancellation, backpressure, batch writes,
  graceful shutdown, and observable worker ownership instead of detached goroutines.
- Keep the API stateless and horizontally scalable; treat PostgreSQL as the initial
  shared bottleneck and defer Kafka, Redis, and microservices until measurements demand them.
- Provide authenticated operations endpoints and development/staging diagnostics, but
  no product or development MCP endpoint in the production runtime.

## Capabilities

### New Capabilities

- `central-api-contracts`: Defines versioned HTTPS contracts, compatibility, limits,
  idempotency, authentication boundaries, and error behavior.
- `catalog-publication`: Imports approved providers and publishes immutable verified
  card and meta catalog snapshots plus manifests.
- `consented-telemetry`: Collects minimal opt-in batches, supports revocation and
  deletion, and produces bounded aggregates without collection payloads.
- `central-data-governance`: Defines PostgreSQL ownership, roles, isolation, migrations,
  retention, backup, restore, and privacy constraints.
- `central-operations`: Defines worker leases, concurrency budgets, observability,
  health, readiness, graceful shutdown, and incident evidence.

### Modified Capabilities

None. The central platform is a new trust boundary and has no archived capabilities.

## Impact

- Adds a separately versioned `central/` Go module with `api` and `worker` commands.
- Adds PostgreSQL migrations, object-storage adapter, OpenAPI schema, deployment
  configuration, load tests, and disaster-recovery procedures.
- The desktop depends only on versioned HTTPS and catalog artifact contracts, never on
  central Go packages or direct database access.
- Requires infrastructure for TLS ingress, PostgreSQL, object storage/CDN, secrets,
  backups, monitoring, and isolated development/staging environments.
