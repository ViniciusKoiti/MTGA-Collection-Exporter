# System Design Test Plan

## Scope and Verdict

This plan evaluates `mtga-system-design.puml` and `mtga-agent-graph-loop.puml`;
it does not claim that the current Python app already has these properties.

| System | Scalability | Security |
| --- | --- | --- |
| Current Python desktop | Suitable for one user and one MTGA process; no central-load concern | Partial: local data, but memory scanning, broad MCP reads, weak auditability |
| Target three-layer design | Plausible to scale horizontally except for PostgreSQL; not proven until load tests pass | Sound boundaries if collection data stays local and development MCP cannot reach production |

## Capacity Hypotheses

Use `daily events = DAU * sessions/day * events/session`. Start with 1.5 sessions
and 40 permitted aggregate events per session.

| Stage | Daily events | Average | Required burst test |
| --- | ---: | ---: | ---: |
| 10k DAU | 600k | 7 events/s | 70 events/s for 15 min |
| 100k DAU | 6M | 70 events/s | 700 events/s for 15 min |

Telemetry is uploaded in idempotent batches. Catalog payloads are immutable,
hash-verified objects served through a CDN; the API only serves their manifest.

## Test Matrix

| ID | System design test | Initial pass criterion |
| --- | --- | --- |
| CON-01 | Wails binding and central API contract tests | Invalid or incompatible payloads fail before release |
| CON-02 | Event schema compatibility | Current and previous client schemas ingest without data loss |
| LOC-01 | Match 10k decks against a 30k-card catalog | p95 <= 1 s, peak memory <= 250 MB on reference PC |
| LOC-02 | Search, filter and ownership-gap operations | p95 <= 100 ms after local indexing |
| API-01 | Manifest and health load test | p95 <= 200 ms, error rate < 0.1% at required burst |
| TEL-01 | Batched telemetry load test | p95 <= 500 ms, no loss, duplicates have no extra effect |
| SOAK-01 | 24-hour API, worker and outbox soak | No unbounded memory, queue or connection growth |
| RES-01 | Central API unavailable for 24 hours | Local collection and deck work remain usable |
| RES-02 | PostgreSQL, CDN or network fault injection | Bounded retries; last valid catalog remains active |
| DAT-01 | Duplicate and out-of-order events | Idempotency and ordering rules preserve aggregates |
| GRF-01 | Cycles, repeated calls and slow tools | Loop stops at every configured budget and deadline |
| GRF-02 | Approval, cancellation and crash recovery | No effect occurs without the exact live approval token |
| SEC-01 | Development MCP environment isolation | Production DSNs and non-allowlisted hosts are rejected |
| SEC-02 | Injection, traversal, SSRF and oversized input | Inputs are rejected and no privileged operation executes |
| SEC-03 | PostgreSQL privilege and secret audit | Separate least-privilege roles; no secret in code or logs |
| SEC-04 | Tenant and row isolation | API roles cannot own tables or bypass RLS; cross-tenant reads fail |
| PRV-01 | Telemetry opt-out and payload inspection | Opt-out sends zero events; forbidden fields are rejected |
| PRV-02 | Consent revocation and deletion | Queue is purged; traceable central data is deleted or anonymized |
| DR-01 | Backup restore and migration rollback drill | RPO <= 24 h, RTO <= 2 h for central metadata |

## Security Abuse Cases

- Treat deck names, card text, logs, catalog content and model output as data, never instructions.
- Verify that raw logs, full collections, credentials and local paths cannot enter telemetry.
- Require authenticated, audited operations access with short-lived credentials.
- Rate-limit by client installation and endpoint; cap request, batch and decompressed sizes.
- Encrypt transport, rotate secrets and scan dependencies plus built artifacts in CI.
- Give the development MCP a read-only default role and an explicit dev/staging allowlist.

## Tooling and Test Layout

Use Go `test` and benchmarks for local services, Playwright for Wails web flows,
k6 for API load, Testcontainers for PostgreSQL contracts, and Toxiproxy for faults.

```text
tests/system/contracts/    API, bindings, event and catalog schemas
tests/system/load/         manifest, telemetry and catalog workloads
tests/system/resilience/   network, database, CDN and process failures
tests/system/security/     trust boundaries, abuse cases and privacy
tests/system/graph/        budgets, approvals, replay and recovery
```

## Release Gates

1. Contracts and deterministic graph tests pass without network or an LLM.
2. Local performance passes on a documented low-spec Windows reference machine.
3. Central load and soak pass at the target stage with PostgreSQL headroom >= 30%.
4. Security, privacy, restore and production-isolation checks have stored evidence.
5. Raise the target stage only from measured p95, saturation and cost data.

PostgreSQL is the expected first shared bottleneck. Start with a stateless API,
batching, indexes and connection-pool limits; add partitioning or a queue only
after measurements show that the simpler design misses a release gate.
