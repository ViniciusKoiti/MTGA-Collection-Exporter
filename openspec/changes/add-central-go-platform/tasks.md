## 1. Central Module And Contracts

- [x] 1.1 Create the separate `central` Go module with `api`, `worker`, and `migrate` commands plus domain, application, ports, and adapter boundaries.
- [x] 1.2 Add pinned toolchain, formatting, vet, static analysis, race, vulnerability, dependency, and 100-line source-file gates.
- [ ] 1.3 Define OpenAPI, manifest, artifact, installation, consent, telemetry, deletion, job, audit, health, and error schemas.
- [ ] 1.4 Add generated-or-handwritten contract verification and prove desktop code imports no central Go package.
- [x] 1.5 Implement strict configuration parsing, environment identity, secret references, safe defaults, and redacted startup diagnostics.

## 2. PostgreSQL Foundation

- [x] 2.1 Add local PostgreSQL containers and forward SQL migrations for catalogs, installations, consent, telemetry, aggregates, deletion, jobs, outbox, and audit.
- [x] 2.2 Create owner, migrator, API, telemetry, worker, operations, and backup roles with grant and denial tests.
- [x] 2.3 Configure `pgxpool` budgets, acquisition/query timeouts, health metrics, and maximum replicas from reserved database capacity.
- [x] 2.4 Add `sqlc` typed queries and repository adapters with transaction, cancellation, and stable error mapping. (Hand-owned typed SQL chosen over sqlc, as the design explicitly allows; contract suites prevent drift.)
- [x] 2.5 Implement the dedicated migration command with lock, compatibility, preflight, verification, and API readiness refusal.
- [x] 2.6 Add repository and RLS contract suites for cross-principal denial, idempotency, retention, and concurrency.

## 3. HTTP API

- [x] 3.1 Implement `net/http` server lifecycle with header, request, idle, shutdown, body, decompression, response, and concurrency limits.
- [x] 3.2 Add request ID, recovery, strict JSON, authentication, authorization, rate policy, observability, and stable error middleware.
- [x] 3.3 Implement public manifest and compatibility endpoints with ETag, immutable cache, bounded in-process cache, and singleflight.
- [x] 3.4 Implement installation enrollment, token hashing, consent receipt, credential rotation, revocation, and deletion endpoints.
- [ ] 3.5 Implement telemetry batch ingestion with atomic schema validation, installation scope, idempotency, sequence, and admission control.
- [ ] 3.6 Implement separately authenticated operations status, catalog, job, deletion, and audit endpoints without generic SQL or mutation tools.
- [ ] 3.7 Add contract, fuzz, timeout, cancellation, unknown-field, oversized, compressed-bomb, authorization, and idempotency tests.

## 4. Catalog Publication

- [ ] 4.1 Implement approved provider registry, adapter contracts, usage-rights checks, rate limits, and rapid disablement.
- [ ] 4.2 Implement streaming fetch, decode, normalize, validate, quarantine, and bounded batch persistence stages.
- [ ] 4.3 Add S3-compatible object storage with immutable keys, compression, SHA-256, upload verification, and cleanup.
- [x] 4.4 Add Ed25519 manifest signing, key IDs, trusted-key rotation, dual-sign transition, and revocation tests.
- [ ] 4.5 Implement staged publication and atomic current-manifest activation with previous-snapshot fallback.
- [ ] 4.6 Add bounded concurrency, cancellation, backpressure, deterministic reduction, duplicate-run ownership, and goroutine-leak tests.
- [ ] 4.7 Publish one signed fixture card/meta catalog and verify download, signature, hash, compatibility, and offline fallback from the desktop harness.

## 5. Consented Telemetry

- [ ] 5.1 Define purpose-versioned consent and an allowlist of bounded event names, attributes, types, and retention classes.
- [ ] 5.2 Add desktop enrollment, durable local outbox, idempotent batch, retry, revocation, and deletion-secret flows behind opt-in.
- [ ] 5.3 Reject raw logs, collections, decks, paths, prompts, credentials, MTGA identities, unknown fields, and oversized values atomically.
- [ ] 5.4 Implement short-lived accepted-event storage, deterministic aggregation, duplicate suppression, and aggregate-only queries.
- [ ] 5.5 Implement revocation, deletion, tombstones, backup-restore reapplication, and non-identifying completion evidence.
- [ ] 5.6 Add privacy golden, opt-out-zero-event, sequence, duplicate, out-of-order, backpressure, retention, and deletion tests.

## 6. Durable Jobs And Go Concurrency

- [x] 6.1 Implement transactional job enqueue, idempotency, `SKIP LOCKED` claim, lease, heartbeat, attempts, retry, and terminal outcomes.
- [x] 6.2 Build one context-owned poller and bounded worker group whose maximum cannot exceed database and provider budgets. (CI proof: internal/worker 10.618s + concurrency 1.051s, run 31912312876)
- [x] 6.3 Implement pipeline helpers using `errgroup.WithContext`, explicit limits, bounded channels, sender cancellation, and deterministic reducers.
- [x] 6.4 Add a concurrency registry documenting owner, parent context, limit source, downstream budget, result path, and shutdown behavior for every goroutine site.
- [x] 6.5 Prohibit detached handler work, goroutine-per-record inserts, unbounded channels, model-selected concurrency, and hidden background loops with architecture tests.
- [ ] 6.6 Add race, leak, blocked-sender, slow-consumer, cancellation, worker-death, lease-expiry, duplicate-effect, and 24-hour soak tests.

## 7. Operations And Governance

- [ ] 7.1 Implement structured `slog` output, OpenTelemetry-compatible traces, and low-cardinality metrics with privacy tests.
- [ ] 7.2 Implement distinct liveness and readiness for schema, storage, pool pressure, queue age, signing, and publication capability.
- [ ] 7.3 Implement signal-driven shutdown that fails readiness, stops claims, cancels producers, drains workers, checkpoints, stops HTTP, and closes pools.
- [ ] 7.4 Define per-table purpose, classification, retention, deletion, backup, restore, legal hold, and owner documentation.
- [ ] 7.5 Integrate external secret references, rotation tests, container non-root user, read-only filesystem, and restricted network policy.
- [ ] 7.6 Add encrypted backup verification and an isolated restore drill meeting RPO, RTO, migrations, tombstones, and smoke checks.
- [ ] 7.7 Add incident runbooks for database saturation, provider failure, signing key compromise, bad catalog, privacy request, and rollback.

## 8. Scale And Release Gates

- [ ] 8.1 Run Stage 1 manifest and telemetry load tests at 10k-DAU assumptions with p95, error, pool, queue, CPU, I/O, and cost evidence.
- [ ] 8.2 Run Stage 2 tests only after Stage 1 and verify 700 events/s burst plus at least 30 percent PostgreSQL headroom.
- [ ] 8.3 Run 24-hour soak and PostgreSQL, network, object-store, provider, signing, and shutdown fault scenarios with bounded resource growth.
- [ ] 8.4 Run CodeQL, dependency, secret, container, migration, API abuse, SSRF, injection, authorization, and production-MCP absence checks.
- [ ] 8.5 Deploy isolated development and staging environments, verify production identity denial from development MCP and harness, and store evidence.
- [ ] 8.6 Map every specification scenario to automation or an operations drill and validate this OpenSpec in strict mode.
- [ ] 8.7 Enable catalog publication first, then telemetry enrollment separately, with feature flags and independent rollback.
