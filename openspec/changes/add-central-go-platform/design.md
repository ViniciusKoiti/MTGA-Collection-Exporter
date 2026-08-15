## Context

The desktop is local-first and keeps collections, logs, saved decks, and graph runs
on the user's device. A central service is still required to ingest approved card and
meta providers, publish compact catalogs, and collect explicitly consented aggregate
product telemetry. It must scale independently from Windows releases and must never
become a hidden remote collection store or a product MCP endpoint.

## Goals / Non-Goals

**Goals:**

- Provide stable HTTPS contracts and immutable verified catalogs.
- Accept minimal idempotent telemetry with revocation and deletion support.
- Scale stateless API replicas while controlling PostgreSQL and worker pressure.
- Use Go concurrency for bounded I/O and CPU pipelines with cancellation and ownership.
- Operate with observable jobs, least-privilege roles, backups, and recovery drills.

**Non-Goals:**

- Uploading raw logs, full collections, deck contents, paths, prompts, or credentials.
- Sharing Go packages across the desktop/central network trust boundary.
- Starting with microservices, Kafka, Redis, Kubernetes, or a workflow SaaS.
- Exposing arbitrary SQL, shell, database, or MCP access in production.
- Treating goroutines as a substitute for capacity limits or durable queues.

## Decisions

### 1. Use a separate modular-monolith Go module

Create `central/go.mod` with `cmd/api`, `cmd/worker`, and `cmd/migrate`. Packages follow
domain, application, ports, and adapters. API and worker are separate processes but
share one codebase and database contracts. The desktop consumes only OpenAPI, signed
manifest, and artifact schemas; it never imports central Go code.

This isolates release and dependency lifecycles without microservice operations. Split
a service only after ownership, scaling, or failure data shows an independent boundary.

### 2. Prefer standard HTTP and explicit contracts

Use `net/http` with middleware for request ID, recovery, authentication, limits,
timeouts, rate policy, and observability. Maintain an OpenAPI document and contract
tests; select code generation only if a spike proves it reduces drift without hiding
error and limit behavior. JSON decoding rejects unknown fields and trailing values.

Configure server header, request, idle, and shutdown timeouts plus bounded body and
decompression sizes. Catalog GETs support ETag and immutable caching. Telemetry POSTs
use installation-scoped authentication and idempotency keys. Operations use separate
OIDC/SSO authentication and are not part of public client credentials.

### 3. Use PostgreSQL deliberately

Use `pgx/v5` and `pgxpool`; use `sqlc` for typed hand-owned SQL. Keep forward SQL
migrations in source and execute them through a dedicated migrator role and command,
never automatically from API replicas. Roles are owner, migrator, API catalog reader,
telemetry writer, worker, operations reader, and backup; runtime roles cannot own
tables or use `BYPASSRLS`.

The initial schema owns sources, snapshots, artifacts, installations, consent receipts,
telemetry batches, short-lived accepted events, aggregates, deletion requests, jobs,
outbox records, and audit evidence. Large ingest uses transactions and `CopyFrom` or
bounded batches, not one concurrent insert per record.

### 4. Publish catalogs outside the API process

Workers write immutable compressed objects to S3-compatible storage, calculate SHA-256,
and sign manifests with Ed25519. Publication is staged: upload object, verify it, commit
metadata, then atomically make the signed manifest current. CDN serves artifacts; the
API serves small manifests and compatibility decisions. A failed publish leaves the
previous snapshot current.

This removes large downloads from API and PostgreSQL capacity. It adds object-store
operations and signing-key management, justified by integrity and scale.

### 5. Use pseudonymous installation consent, not accounts

Initial catalog access is public and rate-limited. Telemetry enrollment creates a
random installation credential and separate deletion secret after explicit consent.
The server stores token hashes, consent purpose/version/time, coarse client version,
and revocation state. It does not require a Wizards account or collect an MTGA identity.

Telemetry schemas allow only bounded product events. IP addresses may be processed at
the edge for abuse prevention but are not persisted in application events. Raw accepted
events expire quickly after aggregation. Revocation blocks new batches; a deletion
request erases or irreversibly anonymizes traceable data and records only non-identifying
proof of completion where legally required.

### 6. Bound every goroutine and propagate context

`net/http` already owns one goroutine per active request. Handlers do not detach work
after a response. Parallel independent calls use `errgroup.WithContext` plus `SetLimit`.
Every started goroutine has a parent context, named owner, result/error path, concurrency
budget, and shutdown behavior. Channels are bounded and represent ownership transfer;
they are not used where a direct call or mutex is clearer.

The catalog pipeline uses bounded stages: provider fetch, streaming decode, normalize,
validate, batch persist, artifact write, and publish. CPU worker count derives from a
configured cap and available CPUs; I/O limits derive from provider and database budgets.
One stable reducer owns ordering. Cancellation stops upstream producers and all senders
select on context, preventing blocked goroutine leaks.

### 7. Keep graph runs sequential; parallelize across work

One graph run advances one committed node at a time for determinism and recovery.
Workers execute multiple leased runs concurrently under a global limit. Parallel nodes
must be declared, read-only or idempotent, use child contexts, and join through a stable
reducer. No model output can raise concurrency or create goroutines.

### 8. Use PostgreSQL jobs before an external broker

A durable jobs table uses transactional enqueue, idempotency key, attempt policy,
`FOR UPDATE SKIP LOCKED`, expiring leases, and heartbeat. One poller owns a bounded
worker group. Queue depth and oldest age create backpressure and readiness signals.

This is operationally simpler than Kafka or Redis and sufficient for catalog and
aggregation jobs. Add a broker only when measured throughput, retention, fan-out, or
cross-service ownership exceeds PostgreSQL headroom.

### 9. Make database capacity the concurrency budget

Set a total connection budget below PostgreSQL capacity and divide it across maximum
API replicas, workers, migrations, and operations reserve. Each process has explicit
pool maximum, acquisition timeout, and query timeout. Worker concurrency cannot exceed
its pool or downstream provider limits. Saturation rejects or delays work; it does not
create more goroutines. Telemetry returns bounded retry guidance when admission closes.

### 10. Use layered caching without Redis initially

Immutable artifacts and manifests use CDN caching. API replicas keep only small bounded
in-process caches and use singleflight for concurrent identical refreshes. PostgreSQL
remains authoritative. Redis is deferred until a measured cross-replica cache or global
rate-limit requirement outweighs its failure and operational cost.

### 11. Design shutdown and observability with the runtime

`signal.NotifyContext` starts shutdown: fail readiness, stop job claims, cancel producers,
drain bounded workers, flush durable checkpoints, shut down HTTP, then close pools.
Every phase has a deadline and incomplete work remains lease-recoverable.

Use structured `slog` records and OpenTelemetry-compatible traces and metrics. Labels
are low-cardinality; payloads, tokens, installation IDs, decks, and run IDs are not metric
labels. Health checks process liveness; readiness checks migrations, essential storage,
pool pressure, queue age, and signing/publication capability without leaking internals.

### 12. Define measurable scale gates

Stage 1 targets 10k DAU and 70 events/s burst; Stage 2 targets 100k DAU and 700 events/s
burst with batching. Require p95 manifest <= 200 ms, telemetry <= 500 ms, errors below
0.1%, 24-hour soak stability, and at least 30% PostgreSQL connection/CPU/I/O headroom.
Scale replicas only within the connection budget and revisit architecture from profiles.

## Trade-off Summary

| Decision | Benefit | Cost / trigger to revisit |
| --- | --- | --- |
| Modular monolith | Simple deployment and transactions | Split when ownership or scaling diverges |
| `net/http` | Small dependency surface | Add router only for measured complexity |
| `pgx` + `sqlc` | PostgreSQL features and typed SQL | More explicit SQL ownership |
| PostgreSQL jobs | Durable and operationally simple | Broker when fan-out/throughput exceeds headroom |
| CDN artifacts | Cheap scalable downloads | Signing and publication complexity |
| Bounded goroutines | Throughput with predictable resources | Requires budgets and backpressure design |
| No Redis initially | Fewer failure modes | Add for measured shared-cache/global-limit need |

## Risks / Trade-offs

- **PostgreSQL becomes saturated** -> Admission control, pool budgets, batching, indexes,
  query budgets, load tests, and measured partitioning before new infrastructure.
- **Goroutine or channel leaks** -> Parent contexts, bounded buffers, ownership rules,
  race/leak/soak tests, and shutdown assertions.
- **Telemetry becomes identifying** -> Schema allowlist, short retention, consent,
  deletion, privacy review, and no collection payload.
- **Signing key compromise** -> External secret custody, rotation, key IDs, dual-sign
  transition, client trust set, and revocation procedure.
- **PostgreSQL jobs create contention** -> Separate indexes/tables, short claims, leases,
  metrics, and broker migration criteria.
- **One module grows too broad** -> Package boundaries, dependency tests, 100-line Go
  files, and extraction only with measured independent ownership.

## Migration Plan

1. Add central module, configuration, HTTP skeleton, migrations, local containers,
   role tests, health, graceful shutdown, and CI.
2. Publish a signed fixture catalog through local object storage and verify it locally.
3. Add one approved provider worker and staging publication with rollback.
4. Add consent enrollment and telemetry ingestion behind a disabled client feature.
5. Add aggregation, revocation, deletion, operations, backups, and restore drill.
6. Pass Stage 1 load, soak, fault, privacy, and security gates before production.
7. Roll back clients to the last compatible manifest and disable telemetry enrollment;
   database migrations remain forward-only with restore-based disaster recovery.

## Open Questions

- Which cloud and object-store implementation will host the first environment?
- What exact raw-event retention is legally and operationally justified?
- Which OIDC provider protects operations access?
- What measured threshold triggers partitioning or an external broker?

## External References

- Go pipeline cancellation and bounded parallelism: https://go.dev/blog/pipelines
- Go context contract: https://pkg.go.dev/context
- Bounded errgroup: https://pkg.go.dev/golang.org/x/sync/errgroup
- PostgreSQL pool: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool
- Typed SQL generation: https://docs.sqlc.dev/en/latest/
