# Incident runbooks

OpenSpec `add-central-go-platform`, task 7.7. Each runbook lists
detection, immediate action, recovery and evidence. All operator reads
go through the operations plane (`/ops/v1/*`, operations scope) — no
generic SQL in an incident, ever.

## 1. Database saturation

- **Detect**: readiness fails on `pool`; `/ops/v1/status` shows
  `pool_in_use` pinned at `pool_max`, queue depth and oldest-job age
  climbing.
- **Act**: scale API replicas DOWN toward `Capacity.MaxAPIReplicas`
  (the budget exists to be respected, not raised), pause the worker
  poller (its context cancel stops claims; leases expire safely).
- **Recover**: watch pool pressure fall; resume the poller; jobs
  reclaim automatically via lease expiry — no manual requeue.
- **Evidence**: status snapshots before/after, audit entries.

## 2. Provider failure or misbehavior

- **Act**: `Registry.Disable(provider)` — the guard stops the very
  next job mid-run; no deploy needed. Publication falls back to the
  current manifest (nothing activates on failure).
- **Recover**: verify upstream, `Enable`, run one staged publication
  and compare counts against the last good snapshot before activation.
- **Evidence**: quarantine records, disabled window in audit.

## 3. Signing key compromise

- **Act**: remove the key ID from the desktop trusted set (dual-sign
  transition path from task 4.4), disable publication (readiness probe
  `signing` goes red on purpose), rotate to the standby key ID.
- **Recover**: re-sign the current snapshot with the new key,
  dual-sign until fleet telemetry shows the old key unused, then
  revoke it.
- **Evidence**: key IDs and timestamps in audit; revocation test run.

## 4. Bad catalog published

- **Act**: `RollbackToPrevious` — one transaction, previous snapshot
  becomes current; clients revalidate via ETag within the cache TTL.
- **Recover**: quarantine the bad snapshot's source data, fix
  validation so the same defect cannot pass again (the fix ships with
  a regression test), republish staged.
- **Evidence**: both artifact IDs, rollback audit row, the new test.

## 5. Privacy request (deletion)

- **Act**: the self-serve path is `POST /v1/installations/deletion`
  (secret-gated). For a manual/legal request, file the deletion row
  and revoke through the same repository function — never ad-hoc SQL.
- **Recover**: purge pseudonymous rows per `data-inventory.md`
  retention table; record non-identifying completion proof.
- **Evidence**: `deletion_requests` row with proof; restore drills
  must reapply the tombstone.

## 6. Release rollback

- **Act**: feature flags first (catalog publication and telemetry
  enrollment roll back independently, task 8.7); then deploy the
  previous binary. The migrator refuses schemas ahead of the binary
  ("ahead of this binary" preflight), so roll the schema forward-only
  — never `DROP` in an incident.
- **Recover**: readiness green on all six capabilities before
  re-enabling flags.
- **Evidence**: flag timeline, deploy IDs, readiness transcript.
