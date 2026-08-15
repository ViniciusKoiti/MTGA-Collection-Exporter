## 1. Architecture Prerequisites

- [ ] 1.1 Revise the companion change so its new MCP is development-only and the current Python MCP is documented as legacy compatibility.
- [ ] 1.2 Verify the companion Go domain, application, policy, and storage ports exist before adding workflow adapters.
- [ ] 1.3 Create an activity inventory mapping every external command and scheduled job to a graph kind or a documented pure-query exemption.
- [x] 1.4 Add package dependency tests that keep workflow definitions independent from Wails, SQLite, PostgreSQL, MCP, and model SDKs.

## 2. Graph Runtime Contracts

- [x] 2.1 Define graph identity, node, transition, outcome, limit, recovery, run, step, and typed-error contracts.
- [x] 2.2 Implement registry validation for identities, reachability, terminal outcomes, transitions, and graph-version uniqueness.
- [x] 2.3 Implement the engine that executes nodes and resolves only compiled transitions from typed outcomes.
- [x] 2.4 Enforce step, tool-call, repeated-call, payload-size, active-deadline, and cancellation limits with stable outcomes.
- [x] 2.5 Integrate the policy and approval ports so exact effect previews pause and resume a run safely.
- [x] 2.6 Add in-memory run, approval, outbox, clock, and ID adapters for runtime unit tests.
- [ ] 2.7 Add an activity registry and architecture test that rejects external commands or jobs which bypass a graph.
- [x] 2.8 Add table-driven and race tests for valid graphs, invalid definitions, policy rejection, limits, cancellation, and concurrency.

## 3. Persistence And Recovery

- [x] 3.1 Add SQLite migrations for graph identities, runs, steps, approvals, leases, idempotency keys, events, and outbox records.
- [x] 3.2 Implement transactional checkpoint and outbox writes with optimistic run versions.
- [x] 3.3 Implement bounded leases, restart recovery, incompatible-version handling, and abandoned-run cleanup.
- [x] 3.4 Implement idempotent effect dispatch and acknowledgement around every pre- and post-effect crash point.
- [x] 3.5 Create a reusable run-store contract suite and execute it against in-memory and SQLite adapters.
- [ ] 3.6 Add a development/staging PostgreSQL run-store adapter and pass the same contract suite with Testcontainers.
- [ ] 3.7 Add migration, backup, rollback, corrupt-checkpoint, dual-worker, and crash-recovery integration tests.

## 4. Application Workflow Graphs

- [ ] 4.1 Implement `collection-sync` from source detection through validation, snapshot commit, and compatibility projections.
- [ ] 4.2 Implement `meta-deck-recommendation` from snapshot and catalog loading through matching, ranking, gaps, and evidence.
- [ ] 4.3 Implement `approved-export` with immutable preview, expiring approval, idempotent write, and completion evidence.
- [ ] 4.4 Implement `telemetry-flush` with consent check, minimization, batching, retry, acknowledgement, and local-only fallback.
- [ ] 4.5 Implement `development-scenario` to arrange fixtures, run one target graph, assert evidence, and produce a report.
- [ ] 4.6 Route Wails commands and background jobs through the activity registry and publish typed progress events.
- [ ] 4.7 Add golden graph tests for success, validation failure, retry, approval, cancellation, offline mode, and stale catalog paths.

## 5. Programmatic Development Harness

- [x] 5.1 Define versioned Go and JSON scenario schemas with strict unknown-field, size, and fixture validation.
- [ ] 5.2 Build the harness composition profile around the production engine, registry, nodes, migrations, and activity entry points.
- [ ] 5.3 Implement controlled clock, ID, planner, provider, process, central API, and effect adapters through production ports.
- [x] 5.4 Implement latency, timeout, malformed-response, disconnect, and checkpoint-crash fault decorators.
- [ ] 5.5 Implement ordered transition, outcome, attempt, effect, budget, database, redaction, and forbidden-behavior assertions.
- [ ] 5.6 Add sanitized small, large, malformed, duplicate-printing, unknown-card, and nearly-buildable-deck fixtures.
- [ ] 5.7 Enforce deterministic, integration, end-to-end, staging, and provider-smoke trust profiles and network allowlists.
- [x] 5.8 Add repeatability tests proving identical normalized evidence for the same deterministic scenario and seed.

## 6. Workflow Observability

- [x] 6.1 Define the versioned event envelope and an allowlist of bounded, non-private attributes.
- [x] 6.2 Implement transactional event journaling plus paginated run checkpoint and timeline queries.
- [x] 6.3 Implement sink-level forbidden-field rejection and tests for logs, collection data, paths, credentials, prompts, and payloads.
- [x] 6.4 Implement low-cardinality counters and durations without run, user, card, or deck labels.
- [x] 6.5 Implement dry-run replay using pinned graph versions and sanitized fixture references without effect dispatch.
- [x] 6.6 Implement age and size retention that preserves active checkpoints and authoritative domain snapshots.
- [x] 6.7 Implement a redacted diagnostic bundle and compare it against privacy golden files.

## 7. Development MCP

- [ ] 7.1 Add a separate stdio `dev-mcp` command and exclude it from every product packaging manifest.
- [ ] 7.2 Implement pre-connection environment classification that denies production and ambiguous hosts, DSNs, certificates, and markers.
- [ ] 7.3 Expose only graph listing, scenario execution, run timeline, event validation, and fixture diagnostic tools.
- [ ] 7.4 Enforce read-only inspection by default and a separate scenario-execute capability restricted to isolated namespaces.
- [ ] 7.5 Add request, response, duration, concurrency, pagination, and audit limits without recording prompts.
- [ ] 7.6 Add MCP contract and abuse tests for malformed arguments, arbitrary SQL, shell requests, SSRF, oversized results, and disconnects.

## 8. Quality And Release Gates

- [x] 8.1 Add a CI check limiting manually maintained Go files to 100 physical lines with only documented generated, migration, lockfile, and fixture exclusions.
- [ ] 8.2 Run deterministic workflow, harness, policy, replay, race, and forbidden-field tests without MTGA, network, administrator access, or an LLM.
- [ ] 8.3 Run SQLite and PostgreSQL integration suites with isolated databases and prove production identities are rejected.
- [ ] 8.4 Run Playwright end-to-end scenarios against Wails progress, approval, cancellation, recovery, and error states.
- [ ] 8.5 Run local performance, central load, soak, and fault scenarios against the thresholds in `docs/architecture/system-design-tests.md`.
- [ ] 8.6 Inspect release artifacts to prove the development MCP, test credentials, private fixtures, and development configuration are absent.
- [ ] 8.7 Map every specification scenario to an automated test or a documented Windows-only manual verification.
- [ ] 8.8 Validate the OpenSpec change in strict mode and store CI evidence for all release gates.
